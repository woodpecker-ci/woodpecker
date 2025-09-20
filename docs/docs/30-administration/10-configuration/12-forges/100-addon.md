# Custom

If the forge you are using does not meet the [Woodpecker requirements](../../../92-development/02-core-ideas.md#forges) or your setup is too specific to be included in the Woodpecker core, you can write an addon forge.

:::warning
Addon forges are still experimental. Their implementation can change and break at any time.
:::

:::danger
You must trust the author of the addon forge you are using. They may have access to authentication codes and other potentially sensitive information.
:::

## Usage

To use an addon forge, download the correct addon version. Then, you can add the following to your configuration:

```ini
WOODPECKER_ADDON_FORGE=/path/to/your/addon/forge/file
```

In case you run Woodpecker as container, you probably want to mount the addon binary to `/opt/addons/`.

## List of addon forges

- [Radicle](https://radicle.xyz/): Open source, peer-to-peer code collaboration stack built on Git. Radicle addon for Woodpecker CI can be found at [this repo](https://explorer.radicle.gr/nodes/seed.radicle.gr/rad:z39Cf1XzrvCLRZZJRUZnx9D1fj5ws).

## Developing addon forges

See [Addons](../../../92-development/100-addons.md).
