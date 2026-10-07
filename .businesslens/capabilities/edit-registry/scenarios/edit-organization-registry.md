---
kind: alternative
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
    web: {place: 'web::signed-in::organization-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Registry of the Organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: registry
    effect: changes
    facts: [Username, Password]
  contexts:
    web: {place: 'web::signed-in::organization-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Edit a organization registry

## Trigger

A registry of a team organization must change.

## Outcome

Pipelines started afterwards receive the new registry.
