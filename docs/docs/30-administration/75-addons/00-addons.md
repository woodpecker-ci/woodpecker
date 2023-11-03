# Addons

To adapt Woodpecker to your needs beyond the [configuration](../10-server-config.md), Woodpecker has its own **addon** system, built ontop of [Go's internal plugin system](https://go.dev/pkg/plugin).

Addons can be used for:

- Forges
- Agent backends

## Restrictions

Addons are restricted by how Go plugins work. This includes the following restrictions:
- only supported on Linux, FreeBSD and macOS
- addons must have been built for the correct Woodpecker version. If an addon is not provided specifically for this version, you likely won't be able to use it.

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

You can list multiple addons, Woodpecker will automatically determine their type. If you specify multiple addons with the same type, only the first one will be used.
