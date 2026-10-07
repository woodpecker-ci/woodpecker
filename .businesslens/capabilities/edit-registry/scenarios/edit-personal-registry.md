---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User changes a Registry and saves it
  kind: actor
  actor: user
  entities:
  - entity: registry
    effect: reads
    facts: [Address]
  contexts:
    web: {place: 'web::signed-in::user-registries'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Registry of the User's personal organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: registry
    effect: changes
    facts: [Username, Password]
  contexts:
    web: {place: 'web::signed-in::user-registries'}
    api: {place: 'api::signed-in'}
---

# Edit a personal registry

## Trigger

A registry of their personal organization must change.

## Outcome

Pipelines started afterwards receive the new registry.
