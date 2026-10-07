---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Delete on a Registry and confirms
  kind: actor
  actor: user
  entities:
  - entity: registry
    effect: reads
    facts: [Address]
  contexts:
    web: {place: 'web::signed-in::repository-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Registry of the Repository
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - {entity: registry, effect: removes}
  contexts:
    web: {place: 'web::signed-in::repository-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Delete a repository registry

## Trigger

A registry of a repository is no longer needed.

## Outcome

Pipelines no longer receive the registry.
