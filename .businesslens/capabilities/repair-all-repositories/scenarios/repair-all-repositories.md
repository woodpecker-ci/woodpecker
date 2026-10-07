---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator chooses Repair all
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::administration::all-repositories'}
    api: {place: 'api::administration'}
- text: The Product repairs every enabled Repository
  kind: product
  actor: administrator
  entities:
  - entity: repository
    effect: changes
    facts: [Full name, Forge link, Default branch, Webhook]
  contexts:
    web: {place: 'web::administration::all-repositories'}
    api: {place: 'api::administration'}
---

# Repair all repositories

## Trigger

The server's address changed or webhooks were lost.

## Outcome

Every enabled repository has a working webhook.
