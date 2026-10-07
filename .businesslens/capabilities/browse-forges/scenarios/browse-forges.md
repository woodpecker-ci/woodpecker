---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator opens the forges
  kind: actor
  actor: administrator
  entities:
  - entity: forge
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::administration::forges'}
    api: {place: 'api::administration'}
- text: The Product lists every Forge with its type and URL
  kind: product
  actor: administrator
  entities:
  - entity: forge
    effect: reads
    facts: [Forge type, URL]
  contexts:
    web: {place: 'web::administration::forges'}
    api: {place: 'api::administration'}
---

# Browse forges

## Trigger

An administrator reviews the server's forge connections.

## Outcome

The administrator sees each forge.
