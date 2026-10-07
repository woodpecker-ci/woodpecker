---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator chooses the forge type, enters its URL, OAuth client and options and chooses Add
  kind: actor
  actor: administrator
  entities:
  - entity: forge
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::administration::forge'}
    api: {place: 'api::administration'}
- text: The Product creates the Forge
  kind: product
  actor: administrator
  entities:
  - entity: forge
    effect: creates
    facts: [Forge type, URL, OAuth client ID, OAuth client secret, OAuth host, Skip SSL verification, Allowed organizations, Advanced options]
  contexts:
    web: {place: 'web::administration::forge'}
    api: {place: 'api::administration'}
---

# Add a forge

## Trigger

Repositories on another forge are to be built.

## Outcome

People can sign in through the new forge.
