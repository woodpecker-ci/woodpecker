---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator chooses Delete user and confirms
  kind: actor
  actor: administrator
  entities:
  - entity: account
    effect: reads
    facts: [Login]
  - entity: user
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::administration::users'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product deletes the Account and the repositories it owns
  kind: product
  actor: administrator
  entities:
  - {entity: account, effect: removes}
  - {entity: organization, effect: removes, with: account}
  - {entity: repository, effect: removes, from: Enabled, with: organization}
  contexts:
    web: {place: 'web::administration::users'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Delete a user

## Trigger

A person leaves.

## Outcome

The person and their repositories are gone from Woodpecker.
