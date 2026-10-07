---
kind: alternative
routes: {cli: CLI}
steps:
- text: The User changes a Secret and saves it
  kind: actor
  actor: user
  entities:
  - entity: secret
    effect: reads
    facts: [Name]
  contexts:
    cli: {place: 'cli::signed-in'}
- text: The Product saves the Secret of the Repository
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: secret
    effect: changes
    facts: [Value, Available for plugins, Available at events]
  contexts:
    cli: {place: 'cli::signed-in'}
---

# Edit a repository secret from the CLI

## Trigger

A secret of a repository must change.

## Outcome

Pipelines started afterwards receive the new secret.
