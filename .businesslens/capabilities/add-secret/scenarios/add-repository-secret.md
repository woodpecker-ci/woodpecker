---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User enters a name, value, note, events and plugins and chooses Add
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::repository-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product creates the Secret for the Repository
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: secret
    effect: creates
    facts: [Name, Value, Note, Available for plugins, Available at events]
  contexts:
    web: {place: 'web::signed-in::repository-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Add a repository secret

## Trigger

Pipelines of a repository need a secret.

## Outcome

Pipelines of a repository receive the secret.

## Edge cases

- Names are unique per repository, organization or server; saving a duplicate fails.
- The CLI takes no note.
