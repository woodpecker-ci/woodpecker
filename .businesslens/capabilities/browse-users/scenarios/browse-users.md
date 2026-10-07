---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator opens the users
  kind: actor
  actor: administrator
  entities:
  - entity: user
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::administration::users'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product lists every Account
  kind: product
  actor: administrator
  entities:
  - entity: account
    effect: reads
    facts: [Login, Email, Avatar URL, Admin]
  contexts:
    web: {place: 'web::administration::users'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Browse users

## Trigger

An administrator reviews who has an account.

## Outcome

The administrator sees every account.
