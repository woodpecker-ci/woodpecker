# Core ideas

- A configuration (e.g. of a pipeline) should never be [turing complete](https://en.wikipedia.org/wiki/Turing_completeness) (We have agents to exec things 🙂).
- If possible, follow the [KISS principle](https://en.wikipedia.org/wiki/KISS_principle).
- What is used most often should be default.
- Keep different topics separated, so you can write plugins, port new ideas ... more easily, see [Architecture](./05-architecture.md).

## Addons and extensions

If you are wondering whether your contribution will be accepted to be merged in the Woodpecker core, or whether it's better to write an
[addon forge](../30-administration/10-configuration/12-forges/100-addon.md), [extension](../30-administration/10-configuration/10-server.md#external-configuration-api) or an
[external custom backend](../30-administration/10-configuration/11-backends/50-custom.md), please check these points:

- Is your change very specific to your setup and unlikely to be used by anyone else?
- Does your change violate the [guidelines](#guidelines)?

Both should be false when you open a pull request to get your change into the core repository.

### Guidelines

#### Forges

A new forge must support these features:

- OAuth2
- Webhooks
