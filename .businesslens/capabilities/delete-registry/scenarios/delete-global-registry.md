---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator chooses Delete on a Registry and confirms
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
- text: The Product deletes the Registry of the whole server
  kind: product
  actor: administrator
  entities:
  - {entity: registry, effect: removes}
  contexts:
    web: {place: 'web::administration::global-registries'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Delete a global registry

## Trigger

A registry of the whole server is no longer needed.

## Outcome

Pipelines no longer receive the registry.
