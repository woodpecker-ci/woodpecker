---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator changes a Registry and saves it
  kind: actor
  actor: administrator
  entities:
  - entity: registry
    effect: reads
    facts: [Address]
  contexts:
    web: {place: 'web::administration::global-registries'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product saves the Registry of the whole server
  kind: product
  actor: administrator
  entities:
  - entity: registry
    effect: changes
    facts: [Username, Password]
  contexts:
    web: {place: 'web::administration::global-registries'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Edit a global registry

## Trigger

A registry of the whole server must change.

## Outcome

Pipelines started afterwards receive the new registry.
