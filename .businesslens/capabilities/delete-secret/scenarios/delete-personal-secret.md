---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User chooses Delete on a Secret and confirms
  kind: actor
  actor: user
  entities:
  - entity: secret
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::user-secrets'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Secret of the User's personal organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - {entity: secret, effect: removes}
  contexts:
    web: {place: 'web::signed-in::user-secrets'}
    api: {place: 'api::signed-in'}
---

# Delete a personal secret

## Trigger

A secret of their personal organization is no longer needed.

## Outcome

Pipelines no longer receive the secret.
