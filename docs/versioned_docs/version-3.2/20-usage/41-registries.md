# Registries

Woodpecker provides the ability to add container registries in the settings of your repository. Adding a registry allows you to authenticate and pull private images from a container registry when using these images as a step inside your pipeline. Using registry credentials can also help you avoid rate limiting when pulling images from public registries.

## Images from private registries

You must provide registry credentials in the UI in order to pull private container images defined in your YAML configuration file.

These credentials are never exposed to your steps, which means they cannot be used to push, and are safe to use with pull requests, for example. Pushing to a registry still requires setting credentials for the appropriate plugin.

Example configuration using a private image:

```diff
 steps:
   - name: build
+    image: gcr.io/custom/golang
     commands:
       - go build
       - go test
```

Woodpecker matches the registry hostname to each image in your YAML. If the hostnames match, the registry credentials are used to authenticate to your registry and pull the image. Note that registry credentials are used by the Woodpecker agent and are never exposed to your build containers.

Example registry hostnames:

- Image `gcr.io/foo/bar` has hostname `gcr.io`
- Image `foo/bar` has hostname `docker.io`
- Image `qux.com:8000/foo/bar` has hostname `qux.com:8000`

Example registry hostname matching logic:

- Hostname `gcr.io` matches image `gcr.io/foo/bar`
- Hostname `docker.io` matches `golang`
- Hostname `docker.io` matches `library/golang`
- Hostname `docker.io` matches `bradrydzewski/golang`
- Hostname `docker.io` matches `bradrydzewski/golang:latest`

## Global registry support

To make a private registry globally available, check the [server configuration docs](../30-administration/10-server-config.md#global-registry-setting).

## GCR registry support

For specific details on configuring access to Google Container Registry, please view the docs [here](https://cloud.google.com/container-registry/docs/advanced-authentication#using_a_json_key_file).

## Local Images

:::warning
For this, privileged rights are needed only available to admins. In addition, this only works when using a single agent.
:::

It's possible to build a local image by mounting the docker socket as a volume.

With a `Dockerfile` at the root of the project:

```yaml
steps:
  - name: build-image
    image: docker
    commands:
      - docker build --rm -t local/project-image .
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock

  - name: build-project
    image: local/project-image
    commands:
      - ./build.sh
```
