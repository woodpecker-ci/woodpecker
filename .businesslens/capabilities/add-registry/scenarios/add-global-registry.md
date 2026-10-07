---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator enters an address, username and password and chooses Add
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::administration::global-registries'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product creates the Registry for the whole server
  kind: product
  actor: administrator
  entities:
  - entity: registry
    effect: creates
    facts: [Address, Username, Password]
  contexts:
    web: {place: 'web::administration::global-registries'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Add a global registry

## Trigger

Pipelines of the whole server need a registry.

## Outcome

Pipelines of the whole server receive the registry.

## Edge cases

- Addresses are unique per repository, organization or server; saving a duplicate fails.
