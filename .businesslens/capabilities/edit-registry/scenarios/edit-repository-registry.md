---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User changes a Registry and saves it
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
- text: The Product saves the Registry of the Repository
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: registry
    effect: changes
    facts: [Username, Password]
  contexts:
    web: {place: 'web::signed-in::repository-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Edit a repository registry

## Trigger

A registry of a repository must change.

## Outcome

Pipelines started afterwards receive the new registry.
