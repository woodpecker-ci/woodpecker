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
    web: {place: 'web::signed-in::organization-secrets'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Secret of the Organization
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
    web: {place: 'web::signed-in::organization-secrets'}
    api: {place: 'api::signed-in'}
---

# Edit a organization secret

## Trigger

A secret of a team organization must change.

## Outcome

Pipelines started afterwards receive the new secret.
