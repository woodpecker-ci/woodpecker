---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator opens the server's secrets
  kind: actor
  actor: administrator
  entities:
  - entity: secret
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::administration::global-secrets'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product lists the Secrets of the whole server
  kind: product
  actor: administrator
  entities:
  - entity: secret
    effect: reads
    facts: [Name, Note, Available for plugins, Available at events]
  contexts:
    web: {place: 'web::administration::global-secrets'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Browse the server's secrets

## Trigger

An administrator checks what every pipeline receives.

## Outcome

The administrator sees the server's secrets.
