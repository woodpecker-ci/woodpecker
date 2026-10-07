---
kind: primary
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
    web: {place: 'web::signed-in::repository-secrets'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Secret of the Repository
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: secret
    effect: changes
    facts: [Value, Note, Available for plugins, Available at events]
  contexts:
    web: {place: 'web::signed-in::repository-secrets'}
    api: {place: 'api::signed-in'}
---

# Edit a repository secret

## Trigger

A secret of a repository must change.

## Outcome

Pipelines started afterwards receive the new secret.
