---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User chooses Delete on a Registry and confirms
  kind: actor
  actor: user
  entities:
  - entity: registry
    effect: reads
    facts: [Address]
  contexts:
    web: {place: 'web::signed-in::user-registries'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Registry of the User's personal organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - {entity: registry, effect: removes}
  contexts:
    web: {place: 'web::signed-in::user-registries'}
    api: {place: 'api::signed-in'}
---

# Delete a personal registry

## Trigger

A registry of their personal organization is no longer needed.

## Outcome

Pipelines no longer receive the registry.
