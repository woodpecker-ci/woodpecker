---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator chooses Delete forge and confirms
  kind: actor
  actor: administrator
  entities:
  - entity: forge
    effect: reads
    facts: [URL]
  contexts:
    web: {place: 'web::administration::forges'}
    api: {place: 'api::administration'}
- text: The Product deletes the Forge
  kind: product
  actor: administrator
  entities:
  - {entity: forge, effect: removes}
  contexts:
    web: {place: 'web::administration::forges'}
    api: {place: 'api::administration'}
---

# Delete a forge

## Trigger

A forge is no longer used.

## Outcome

People can no longer sign in through the forge.
