# Core ideas

- A (e.g. pipeline) configuration should never be [turing complete](https://en.wikipedia.org/wiki/Turing_completeness) (We have agents to exec things 🙂).
- If possible, follow the [KISS principle](https://en.wikipedia.org/wiki/KISS_principle).
- What is used most should be default.
- Keep different topics separated, so you can write plugins, port new ideas ... more easily, see [Architecture](./05-architecture.md).

## Addons and extensions

If you wonder whether your contribution will be accepted to be merged into our core or it's better to write it as an
[addon](../30-administration/75-addons/00-overview.md), [extension](../30-administration/100-external-configuration-api.md) or an
[external custom backend](../30-administration/22-backends/50-custom-backends.md), please check these points:

- Is your change very specific to your setup and won't likely be used by anybody else?
- Does your change violate the [guidelines](#guidelines)?

Both should be false if you open a pull request to get your change into the core repository.

### Guidelines

#### Forges

A new forge must support these features:

- OAuth2
- Webhooks
