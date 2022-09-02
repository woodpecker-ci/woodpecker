# Example plugin

This provides a brief tutorial for creating a Woodpecker webhook plugin, using simple shell scripting, to make an http requests during the build pipeline.

## What end users will see

The below example demonstrates how we might configure a webhook plugin in the YAML file:

```yaml
pipeline:
  webhook:
    image: foo/webhook
    settings:
      url: http://foo.com
      method: post
      body: |
        hello world
```

## Write the logic

Create a simple shell script that invokes curl using the YAML configuration parameters, which are passed to the script as environment variables in uppercase and prefixed with `PLUGIN_`.

```bash
#!/bin/sh

curl \
  -X ${PLUGIN_METHOD} \
  -d ${PLUGIN_BODY} \
  ${PLUGIN_URL}
```

## Package it

Create a Dockerfile that adds your shell script to the image, and configures the image to execute your shell script as the main entrypoint.

```dockerfile
FROM alpine
ADD script.sh /bin/
RUN chmod +x /bin/script.sh
RUN apk -Uuv add curl ca-certificates
ENTRYPOINT /bin/script.sh
```

Build and publish your plugin to the Docker registry. Once published your plugin can be shared with the broader Woodpecker community.

```nohighlight
docker build -t foo/webhook .
docker push foo/webhook
```

Execute your plugin locally from the command line to verify it is working:

```nohighlight
docker run --rm \
  -e PLUGIN_METHOD=post \
  -e PLUGIN_URL=http://foo.com \
  -e PLUGIN_BODY="hello world" \
  foo/webhook
```
