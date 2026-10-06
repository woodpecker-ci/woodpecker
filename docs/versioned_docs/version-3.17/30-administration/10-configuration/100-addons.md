# Addons

Addons can be used to extend the Woodpecker server. Currently, they can be used for forges and the log service.

:::warning
Addon forges are still experimental. Their implementation can change and break at any time.
:::

:::danger
You must trust the author of the addon forge you are using. They may have access to authentication codes and other potentially sensitive information.
:::

## Usage

To use an addon forge, download the correct addon version.

### Forge

Use this in your `.env`:

```ini
WOODPECKER_ADDON_FORGE=/path/to/your/addon/forge/file
```

In case you run Woodpecker as container, you probably want to mount the addon binary to `/opt/addons/`.

#### List of addon forges

- [Radicle](https://radicle.xyz/): Open source, peer-to-peer code collaboration stack built on Git. Radicle addon for Woodpecker CI can be found at [this repo](https://explorer.radicle.gr/nodes/seed.radicle.gr/rad:z39Cf1XzrvCLRZZJRUZnx9D1fj5ws).

### Log

Use this in your `.env`:

```ini
WOODPECKER_LOG_STORE=addon
WOODPECKER_LOG_STORE_FILE_PATH=/path/to/your/addon/forge/file
```

## Developing addon forges

See [Addons](../../92-development/100-addons.md).
