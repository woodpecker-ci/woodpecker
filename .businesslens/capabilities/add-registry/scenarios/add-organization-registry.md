---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User enters an address, username and password and chooses Add
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::organization-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product creates the Registry for the Organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: registry
    effect: creates
    facts: [Address, Username, Password]
  contexts:
    web: {place: 'web::signed-in::organization-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Add a organization registry

## Trigger

Pipelines of a team organization need a registry.

## Outcome

Pipelines of a team organization receive the registry.

## Edge cases

- Addresses are unique per repository, organization or server; saving a duplicate fails.
