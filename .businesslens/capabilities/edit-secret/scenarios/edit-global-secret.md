---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The Administrator changes a Secret and saves it
  kind: actor
  actor: administrator
  entities:
  - entity: secret
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::administration::global-secrets'}
    api: {place: 'api::administration'}
- text: The Product saves the Secret of the whole server
  kind: product
  actor: administrator
  entities:
  - entity: secret
    effect: changes
    facts: [Value, Note, Available for plugins, Available at events]
  contexts:
    web: {place: 'web::administration::global-secrets'}
    api: {place: 'api::administration'}
---

# Edit a global secret

## Trigger

A secret of the whole server must change.

## Outcome

Pipelines started afterwards receive the new secret.
