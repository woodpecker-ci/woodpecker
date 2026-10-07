---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator opens the organizations
  kind: actor
  actor: administrator
  entities:
  - entity: organization
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::administration::organizations'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product lists every Organization
  kind: product
  actor: administrator
  entities:
  - entity: organization
    effect: reads
    facts: [Name, Personal]
  contexts:
    web: {place: 'web::administration::organizations'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Browse organizations

## Trigger

An administrator reviews the organizations on the server.

## Outcome

The administrator sees every organization.
