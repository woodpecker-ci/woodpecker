---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator chooses Delete on a Secret and confirms
  kind: actor
  actor: administrator
  entities:
  - entity: secret
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::administration::global-secrets'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product deletes the Secret of the whole server
  kind: product
  actor: administrator
  entities:
  - {entity: secret, effect: removes}
  contexts:
    web: {place: 'web::administration::global-secrets'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Delete a global secret

## Trigger

A secret of the whole server is no longer needed.

## Outcome

Pipelines no longer receive the secret.
