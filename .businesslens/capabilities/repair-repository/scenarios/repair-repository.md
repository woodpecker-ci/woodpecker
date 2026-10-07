---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Repair repository
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::signed-in::repository-actions'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product reads the Repository from the forge again and replaces its webhook
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: changes
    facts: [Full name, Forge link, Default branch, Webhook]
  - entity: forge
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repository-actions'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Repair a repository

## Trigger

Events stop arriving, or the repository was renamed or moved on the forge.

## Outcome

The repository receives events again, under its current name.
