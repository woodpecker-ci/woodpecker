---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User enters an address, username and password and chooses Add
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::user-registries'}
    api: {place: 'api::signed-in'}
- text: The Product creates the Registry for the User's personal organization
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
    web: {place: 'web::signed-in::user-registries'}
    api: {place: 'api::signed-in'}
---

# Add a personal registry

## Trigger

Pipelines of their personal organization need a registry.

## Outcome

Pipelines of their personal organization receive the registry.

## Edge cases

- Addresses are unique per repository, organization or server; saving a duplicate fails.
