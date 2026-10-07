---
kind: alternative
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
    web: {place: 'web::signed-in::organization-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Registry of the Organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - {entity: registry, effect: removes}
  contexts:
    web: {place: 'web::signed-in::organization-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Delete a organization registry

## Trigger

A registry of a team organization is no longer needed.

## Outcome

Pipelines no longer receive the registry.
