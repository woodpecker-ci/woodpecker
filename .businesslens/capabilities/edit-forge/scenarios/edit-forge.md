---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator changes a Forge and chooses Save
  kind: actor
  actor: administrator
  entities:
  - entity: forge
    effect: reads
    facts: [URL]
  contexts:
    web: {place: 'web::administration::forge'}
    api: {place: 'api::administration'}
- text: The Product saves the Forge
  kind: product
  actor: administrator
  entities:
  - entity: forge
    effect: changes
    facts: [Forge type, URL, OAuth client ID, OAuth client secret, OAuth host, Skip SSL verification, Allowed organizations, Advanced options]
  contexts:
    web: {place: 'web::administration::forge'}
    api: {place: 'api::administration'}
---

# Edit a forge

## Trigger

A forge moved or its OAuth application changed.

## Outcome

Woodpecker uses the new connection.
