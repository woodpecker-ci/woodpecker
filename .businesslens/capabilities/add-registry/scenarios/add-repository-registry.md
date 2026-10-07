---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User enters an address, username and password and chooses Add
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::repository-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product creates the Registry for the Repository
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: registry
    effect: creates
    facts: [Address, Username, Password]
  contexts:
    web: {place: 'web::signed-in::repository-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Add a repository registry

## Trigger

Pipelines of a repository need a registry.

## Outcome

Pipelines of a repository receive the registry.

## Edge cases

- Addresses are unique per repository, organization or server; saving a duplicate fails.
