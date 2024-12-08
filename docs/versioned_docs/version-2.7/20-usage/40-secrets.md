# Secrets

Woodpecker provides the ability to store named parameters external to the YAML configuration file, in a central secret store. These secrets can be passed to individual steps of the pipeline at runtime.

Woodpecker provides three different levels to add secrets to your pipeline. The following list shows the priority of the different levels. If a secret is defined in multiple levels, will be used following this priorities: Repository secrets > Organization secrets > Global secrets.

1. **Repository secrets**: They are available to all pipelines of an repository.
2. **Organization secrets**: They are available to all pipelines of an organization.
3. **Global secrets**: Can be configured by an instance admin.
   They are available to all pipelines of the **whole** Woodpecker instance and should therefore **only** be used for secrets that are allowed to be read by **all** users.

## Usage

### Use secrets in commands

Secrets are exposed to your pipeline steps and plugins as uppercase environment variables and can therefore be referenced in the commands section of your pipeline,
once their usage is declared in the `secrets` section:

```diff
 steps:
   - name: "step name"
     image: registry/repo/image:tag
     commands:
+      - echo $some_username
+      - echo $SOME_PASSWORD
+    secrets: [ some_username, SOME_PASSWORD ]
```

The environment variables retain their original case, but secret matching is performed in a case-insensitive manner.
In this example, `DOCKER_PASSWORD` would still match even if the secret is named `docker_password`.

### Use secrets in normal steps via environment

You can set an environment value from secrets using the `from_secret` syntax.
So the secret key and environment variable name can differ.

```diff
 steps:
   - name: test
     image: bash
     commands:
       - env | grep OWN
-    secrets: [ some_username, SOME_PASSWORD ]
+    environment:
+      SOME_OWN_DEFINED_VAR:
+        from_secret: some_username
```

### Use secrets in plugins via settings

The `from_secret` syntax also work for settings in any hierarchy.

In this example, the secret named `secret_token` would be passed to the setting named `SURGE_TOKEN`,which will be available in the plugin as environment variable named `PLUGIN_SURGE_TOKEN` (See [plugins](./51-plugins/20-creating-plugins.md#settings) for details).

```diff
 steps:
   - name: deploy-preview:
     image: woodpeckerci/plugin-surge-preview
     settings:
       path: 'docs/build/'
+      surge_token:
+        from_secret: SURGE_TOKEN
```

As settings can have complex structure, the `from_secret` is supported in all of it:

```yaml
steps:
  - name: deploy-test:
    image: plugin-example
    settings:
      path: 'artifacts'
      simple_token:
        from_secret: A_TOKEN
      advanced:
        items:
          - "value1"
          - some:
              from_secret: secret_value
          - "value3"
```

### Note about parameter pre-processing

Please note parameter expressions are subject to pre-processing. When using secrets in parameter expressions they should be escaped.

```diff
 steps:
   - name: "echo password"
     image: bash
     commands:
-      - echo ${some_username}
-      - echo ${SOME_PASSWORD}
+      - echo $${some_username}
+      - echo $${SOME_PASSWORD}
     secrets: [ some_username, SOME_PASSWORD ]
```

### Use in Pull Requests events

Secrets are not exposed to pull requests by default. You can override this behavior by creating the secret and enabling the `pull_request` event type, either in UI or by CLI, see below.

:::note
Please be careful when exposing secrets to pull requests. If your repository is open source and accepts pull requests your secrets are not safe. A bad actor can submit a malicious pull request that exposes your secrets.
:::

## Plugins filter

To prevent abusing your secrets from malicious usage, you can limit a secret to a list of plugins. If enabled they are not available to any other plugin (steps without user-defined commands). If you or an attacker defines explicit commands, the secrets will not be available to the container to prevent leaking them.

![plugins filter](./secrets-plugins-filter.png)

## Adding Secrets

Secrets are added to the Woodpecker in the UI or with the CLI.

### CLI Examples

Create the secret using default settings. The secret will be available to all images in your pipeline, and will be available to all push, tag, and deployment events (not pull request events).

```bash
woodpecker-cli secret add \
  -repository octocat/hello-world \
  -name aws_access_key_id \
  -value <value>
```

Create the secret and limit to a single image:

```diff
 woodpecker-cli secret add \
   -repository octocat/hello-world \
+  -image plugins/s3 \
   -name aws_access_key_id \
   -value <value>
```

Create the secrets and limit to a set of images:

```diff
 woodpecker-cli secret add \
   -repository octocat/hello-world \
+  -image plugins/s3 \
+  -image peloton/woodpecker-ecs \
   -name aws_access_key_id \
   -value <value>
```

Create the secret and enable for multiple hook events:

```diff
 woodpecker-cli secret add \
   -repository octocat/hello-world \
   -image plugins/s3 \
+  -event pull_request \
+  -event push \
+  -event tag \
   -name aws_access_key_id \
   -value <value>
```

Loading secrets from file using curl `@` syntax. This is the recommended approach for loading secrets from file to preserve newlines:

```diff
 woodpecker-cli secret add \
   -repository octocat/hello-world \
   -name ssh_key \
+  -value @/root/ssh/id_rsa
```
