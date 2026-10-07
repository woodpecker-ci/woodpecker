---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator opens the server's registries
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::administration::global-registries'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product lists the Registries of the whole server
  kind: product
  actor: administrator
  entities:
  - entity: registry
    effect: reads
    facts: [Address, Username, Read only]
  contexts:
    web: {place: 'web::administration::global-registries'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Browse the server's registries

## Trigger

An administrator checks what every pipeline receives.

## Outcome

The administrator sees the server's registries.
