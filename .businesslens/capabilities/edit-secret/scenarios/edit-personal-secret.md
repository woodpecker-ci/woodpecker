---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User changes a Secret and saves it
  kind: actor
  actor: user
  entities:
  - entity: secret
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::user-secrets'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Secret of the User's personal organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: secret
    effect: changes
    facts: [Value, Note, Available for plugins, Available at events]
  contexts:
    web: {place: 'web::signed-in::user-secrets'}
    api: {place: 'api::signed-in'}
---

# Edit a personal secret

## Trigger

A secret of their personal organization must change.

## Outcome

Pipelines started afterwards receive the new secret.
