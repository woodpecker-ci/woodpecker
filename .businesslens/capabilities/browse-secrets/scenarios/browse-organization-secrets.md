---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User opens the secrets of an Organization
  kind: actor
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: secret
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::organization-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product lists the Organization's Secrets
  kind: product
  actor: user
  entities:
  - entity: secret
    effect: reads
    facts: [Name, Note, Available for plugins, Available at events]
  - entity: organization
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::organization-secrets'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Browse an organization's secrets

## Trigger

An organization admin checks the organization's secrets.

## Outcome

The admin sees the organization's secrets.
