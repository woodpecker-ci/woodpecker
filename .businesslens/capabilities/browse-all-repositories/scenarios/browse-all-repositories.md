---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator opens all repositories
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::administration::all-repositories'}
    api: {place: 'api::administration'}
- text: The Product lists every Repository of the server
  kind: product
  actor: administrator
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::administration::all-repositories'}
    api: {place: 'api::administration'}
---

# Browse all repositories

## Trigger

An administrator reviews what the server builds.

## Outcome

The administrator sees every repository.
