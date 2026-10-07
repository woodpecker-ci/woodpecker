---
kind: alternative
routes: {cli: CLI}
steps:
- text: The Administrator changes a Secret and saves it
  kind: actor
  actor: administrator
  entities:
  - entity: secret
    effect: reads
    facts: [Name]
  contexts:
    cli: {place: 'cli::administration'}
- text: The Product saves the Secret of the whole server
  kind: product
  actor: administrator
  entities:
  - entity: secret
    effect: changes
    facts: [Value, Available for plugins, Available at events]
  contexts:
    cli: {place: 'cli::administration'}
---

# Edit a global secret from the CLI

## Trigger

A secret of the whole server must change.

## Outcome

Pipelines started afterwards receive the new secret.
