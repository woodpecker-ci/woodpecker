---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator enters a name, value, note, events and plugins and chooses Add
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::administration::global-secrets'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product creates the Secret for the whole server
  kind: product
  actor: administrator
  entities:
  - entity: secret
    effect: creates
    facts: [Name, Value, Note, Available for plugins, Available at events]
  contexts:
    web: {place: 'web::administration::global-secrets'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Add a global secret

## Trigger

Pipelines of the whole server need a secret.

## Outcome

Pipelines of the whole server receive the secret.

## Edge cases

- Names are unique per repository, organization or server; saving a duplicate fails.
- The CLI takes no note.
