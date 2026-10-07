---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User chooses Reset token
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::cli-and-api'}
    api: {place: 'api::signed-in'}
- text: The Product replaces the Account's personal access token and shows the new one
  kind: product
  actor: user
  entities:
  - entity: account
    effect: changes
    facts: [Personal access token]
  contexts:
    web: {place: 'web::signed-in::cli-and-api'}
    api: {place: 'api::signed-in'}
---

# Reset the personal access token

## Trigger

A person wants to revoke their current token.

## Outcome

Only the new token is accepted; the person signs in again on other devices.
