# Creating plugins

Creating a new plugin is simple: Build a Docker container which uses your plugin logic as entrypoint.

## Settings

To allow users to configure the behavior of your plugin, you should use settings.

These are passed to your plugin as uppercase env vars with `PLUGIN_` prefix.
Using a setting like `url` results in `PLUGIN_URL` as env var.

Characters like `-` are converted to an underscore (`_`). `some_String` gets `PLUGIN_SOME_STRING`.
CamelCase is not respected, `anInt` get `PLUGIN_ANINT`.

### Basic settings

Using any basic YAML type (scalar) will be converted into a string:

| Setting              | Environment value            |
| -------------------- | ---------------------------- |
| `some-bool: false`   | `PLUGIN_SOME_BOOL="false"`   |
| `some_String: hello` | `PLUGIN_SOME_STRING="hello"` |
| `anInt: 3`           | `PLUGIN_ANINT="3"`           |

### Complex settings

It's also possible to use complex settings like this:

```yaml
steps:
  plugin:
    image: foo/plugin
    settings:
      complex:
        abc: 2
        list:
          - 2
          - 3
```

Values like this are converted to JSON and then passed to your plugin. In the example above, the env value would be `{"abc": "2", "list": [ "2", "3" ]}`.

### Secrets

Secrets should be passed as settings too. Therefore, users should use [`from_secret`](../40-secrets.md#use-secrets-in-settings).

## Plugin library

For Go, we provide a plugin library you can use to get easy access to internal env vars and your settings. See <https://codeberg.org/woodpecker-plugins/go-plugin>.

## Example plugin

This provides a brief tutorial for creating a Woodpecker webhook plugin, using simple shell scripting, to make HTTP requests during the build pipeline.

### What end users will see

The below example demonstrates how we might configure a webhook plugin in the YAML file:

```yaml
steps:
  webhook:
    image: foo/webhook
    settings:
      url: https://example.com
      method: post
      body: |
        hello world
```

### Write the logic

Create a simple shell script that invokes curl using the YAML configuration parameters, which are passed to the script as environment variables in uppercase and prefixed with `PLUGIN_`.

```bash
#!/bin/sh

curl \
  -X ${PLUGIN_METHOD} \
  -d ${PLUGIN_BODY} \
  ${PLUGIN_URL}
```

### Package it

Create a Dockerfile that adds your shell script to the image, and configures the image to execute your shell script as the main entrypoint.

```dockerfile
FROM alpine
ADD script.sh /bin/
RUN chmod +x /bin/script.sh
RUN apk -Uuv add curl ca-certificates
ENTRYPOINT /bin/script.sh
```

Build and publish your plugin to the Docker registry. Once published, your plugin can be shared with the broader Woodpecker community.

```shell
docker build -t foo/webhook .
docker push foo/webhook
```

Execute your plugin locally from the command line to verify it is working:

```shell
docker run --rm \
  -e PLUGIN_METHOD=post \
  -e PLUGIN_URL=https://example.com \
  -e PLUGIN_BODY="hello world" \
  foo/webhook
```

## Best practices

- Build your plugin for different architectures to allow many users to use them.
  At least, you should support `amd64` and `arm64`.
- Provide binaries for users using the `local` backend.
  These should also be built for different OS/architectures.
- Use [built-in env vars](../50-environment.md#built-in-environment-variables) where possible.
- Do not use any configuration except settings (and internal env vars). This means: Don't require using [`environment`](../50-environment.md) and don't require specific secret names.
- Add a `docs.md` file, listing all your settings and plugin metadata ([example](https://codeberg.org/woodpecker-plugins/plugin-docker-buildx/src/branch/main/docs.md)).
- Add your plugin to the [plugin index](/plugins) using your `docs.md` ([the example above in the index](https://woodpecker-ci.org/plugins/Docker%20Buildx)).
