---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Delete on a Secret and confirms
  kind: actor
  actor: user
  entities:
  - entity: secret
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::organization-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Secret of the Organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - {entity: secret, effect: removes}
  contexts:
    web: {place: 'web::signed-in::organization-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Delete a organization secret

## Trigger

A secret of a team organization is no longer needed.

## Outcome

Pipelines no longer receive the secret.
