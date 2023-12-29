# Addons

:::warning
Addons are still experimental. Their implementation can change and break at any time.
:::

:::danger
You need to trust the author of the addons you use. Depending on their type, addons can access forge authentication codes, your secrets or other sensitive information.
:::

To adapt Woodpecker to your needs beyond the [configuration](../10-server-config.md), Woodpecker has its own **addon** system, built ontop of [Go's internal plugin system](https://go.dev/pkg/plugin).

Addons can be used for:

- Forges
- Agent backends
- Config services
- Secret services
- Environment services
- Registry services

## Restrictions

Addons are restricted by how Go plugins work. This includes the following restrictions:

- only supported on Linux, FreeBSD and macOS
- addons must have been built for the correct Woodpecker version. If an addon is not provided specifically for this version, you likely won't be able to use it.

## Usage

To use an addon, download the addon version built for your Woodpecker version. Then, you can add the following to your configuration:

```ini
WOODPECKER_ADDONS=/path/to/your/addon/file.so
```

In case you run Woodpecker as container, you probably want to mount the addon binaries to `/opt/addons/`.

You can list multiple addons, Woodpecker will automatically determine their type. If you specify multiple addons with the same type, only the first one will be used.

Using an addon always overwrites Woodpecker's internal setup. This means, that a forge addon will be used if specified, no matter what's configured for the forges natively supported by Woodpecker.

### Bug reports

If you experience bugs, please check which component has the issue. If it's the addon, **do not raise an issue in the main repository**, but rather use the separate addon repositories. To check which component is responsible for the bug, look at the logs. Logs from addons are marked with a special field `addon` containing their addon file name.
