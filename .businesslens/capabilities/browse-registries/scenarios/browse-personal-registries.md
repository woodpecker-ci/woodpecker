---
kind: alternative
routes: {web: Web}
steps:
- text: The User opens their registries
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::user-registries'}
- text: The Product lists the Registries of the User's personal organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: registry
    effect: reads
    facts: [Address, Username, Read only]
  contexts:
    web: {place: 'web::signed-in::user-registries'}
---

# Browse personal registries

## Trigger

A person checks their own registries.

## Outcome

The person sees their registries.
