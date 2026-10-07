---
kind: primary
routes:
  web: Web
steps:
- text: The Administrator opens the admin settings
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::administration::info'}
- text: The Product shows the version the server runs and whether a newer release exists
  kind: product
  actor: administrator
  entities:
  - {entity: server-settings, effect: reads, facts: [Version]}
  contexts:
    web: {place: 'web::administration::info'}
---

# See the server version as an administrator

## Trigger

An administrator checks whether the server needs an update.

## Outcome

The administrator knows the running version and whether to update.
