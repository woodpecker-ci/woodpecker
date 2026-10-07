---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The Administrator enters a username, email and avatar URL and chooses Save user
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
- text: The Product creates the Account
  kind: product
  actor: administrator
  entities:
  - entity: account
    effect: creates
    facts: [Login, Email, Avatar URL, Personal access token]
  contexts:
    web: {place: 'web::administration::users'}
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Add a user

## Trigger

A person must exist before they sign in, for example while registration is closed.

## Outcome

The person can sign in with that forge login.

## Edge cases

- An empty or invalid username is refused.
