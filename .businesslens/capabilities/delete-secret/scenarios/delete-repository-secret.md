---
kind: primary
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
    web: {place: 'web::signed-in::repository-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Secret of the Repository
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - {entity: secret, effect: removes}
  contexts:
    web: {place: 'web::signed-in::repository-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Delete a repository secret

## Trigger

A secret of a repository is no longer needed.

## Outcome

Pipelines no longer receive the secret.
