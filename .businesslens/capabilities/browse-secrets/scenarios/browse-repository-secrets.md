---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User opens the secrets of a Repository
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: secret
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repository-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product lists the Repository's Secrets with those of its organization and the server
  kind: product
  actor: user
  entities:
  - entity: secret
    effect: reads
    facts: [Name, Note, Available for plugins, Available at events]
  - entity: organization
    effect: reads
    facts: []
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repository-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Browse a repository's secrets

## Trigger

A person checks which secrets a repository's pipelines receive.

## Outcome

The person sees the secrets without their confidential values.
