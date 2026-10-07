---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User opens the registries of an Organization
  kind: actor
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::organization-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product lists the Organization's Registries
  kind: product
  actor: user
  entities:
  - entity: registry
    effect: reads
    facts: [Address, Username, Read only]
  - entity: organization
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::organization-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Browse an organization's registries

## Trigger

An organization admin checks the organization's registries.

## Outcome

The admin sees the organization's registries.
