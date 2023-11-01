# Addons

To adapt Woodpecker to your needs beyond the [configuration](../10-server-config.md), Woodpecker has its own **addon** system, built ontop of [Go's internal plugin system](https://go.dev/pkg/plugin).

Currently, addons can only be used to integrate forges into Woodpecker that are not supported out-of-the-box.

## Usage

To use an addon, download the addon version built for your Woodpecker version. Then, you can add the following to your configuration:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
+     - WOODPECKER_PLUGIN=/path/to/your/addon/file.so
```

You may need to [mount the addon file as volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to access it from inside the Docker container.
