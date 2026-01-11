# Secrets

Woodpecker provides the ability to store named variables in a central secret store.
These secrets can be securely passed on to individual pipeline steps using the keyword `from_secret`.

There are three different levels of secrets available. If a secret is defined in multiple levels, the following order of priority applies (last wins):

1. **Repository secrets**: Available for all pipelines of a repository.
1. **Organization secrets**: Available for all pipelines of an organization.
1. **Global secrets**: Can only be set by instance administrators.
   Global secrets are available for all pipelines of the **entire** Woodpecker instance and should therefore be used with caution.

In addition to the native integration of secrets, external providers of secrets can also be used by interacting with them directly within pipeline steps. Access to these providers can be configured with Woodpecker secrets, which enables the retrieval of secrets from the respective external sources.

:::warning
Woodpecker can mask secrets from its own secrets store, but it cannot apply the same protection to external secrets. As a result, these external secrets can be exposed in the pipeline logs.
:::

## Usage

You can set a setting or environment value from Woodpecker secrets by using the `from_secret` syntax.

The following example passes a secret called `secret_token` which is stored in an environment variable called `TOKEN_ENV`:

```diff
 steps:
   - name: 'step name'
     image: registry/repo/image:tag
     commands:
+      - echo "The secret is $TOKEN_ENV"
+    environment:
+      TOKEN_ENV:
+        from_secret: secret_token
```

The same syntax can be used to pass secrets to (plugin) settings.
A secret called `secret_token` is assigned to the setting `TOKEN`, which is then available in the plugin as the environment variable `PLUGIN_TOKEN` (see [plugins](./51-plugins/20-creating-plugins.md#settings) for details).
`PLUGIN_TOKEN` is then used internally by the plugin itself and taken into account during execution.

```diff
 steps:
   - name: 'step name'
     image: registry/repo/image:tag
+    settings:
+      TOKEN:
+        from_secret: secret_token
```

### Escape secrets

Please note that parameter expressions are preprocessed, i.e. they are evaluated before the pipeline starts.
If secrets are to be used in expressions, they must be properly escaped (with `$$`) to ensure correct processing.

```diff
 steps:
   - name: docker
     image: docker
     commands:
-      - echo ${TOKEN_ENV}
+      - echo $${TOKEN_ENV}
     environment:
       TOKEN_ENV:
         from_secret: secret_token
```

### Events filter

By default, secrets are not exposed to pull requests.
However, you can change this behavior by creating the secret and enabling the `pull_request` event type.
This can be configured either via the UI or via the CLI.

:::warning
Be careful when exposing secrets for pull requests.
If your repository is public and accepts pull requests from everyone, your secrets may be at risk.
Malicious actors could take advantage of this to expose your secrets or transfer them to an external location.
:::

### Plugins filter

To prevent your secrets from being misused by malicious users, you can restrict a secret to a list of plugins.
If enabled, they are not available to any other plugins.
Plugins have the advantage that they cannot execute arbitrary commands and therefore cannot reveal secrets.

:::tip
If you specify a tag, the filter will take it into account.
However, if the same image appears several times in the list, the least privileged entry will take precedence.
For example, an image without a tag will allow all tags, even if it contains another entry with a tag attached.
:::

![plugins filter](./secrets-plugins-filter.png)

## CLI

In addition to the UI, secrets can also be managed using the CLI.

Create the secret with the default settings.
The secret is available for all images in your pipeline and for all `push`, `tag` and `deployment` events (not for `pull_request` events).

```bash
woodpecker-cli repo secret add \
  --repository octocat/hello-world \
  --name aws_access_key_id \
  --value <value>
```

Create the secret and limit it to a single image:

```diff
 woodpecker-cli secret add \
   --repository octocat/hello-world \
+  --image woodpeckerci/plugin-s3 \
   --name aws_access_key_id \
   --value <value>
```

Create the secrets and limit it to a set of images:

```diff
 woodpecker-cli repo secret add \
   --repository octocat/hello-world \
+  --image woodpeckerci/plugin-s3 \
+  --image woodpeckerci/plugin-docker-buildx \
   --name aws_access_key_id \
   --value <value>
```

Create the secret and enable it for multiple hook events:

```diff
 woodpecker-cli repo secret add \
   --repository octocat/hello-world \
   --image woodpeckerci/plugin-s3 \
+  --event pull_request \
+  --event push \
+  --event tag \
   --name aws_access_key_id \
   --value <value>
```

Secrets can be loaded from a file using the syntax `@`.
This method is recommended for loading secrets from a file, as it ensures that line breaks are preserved (this is important for SSH keys, for example):

```diff
 woodpecker-cli repo secret add \
   -repository octocat/hello-world \
   -name ssh_key \
+  -value @/root/ssh/id_rsa
```
