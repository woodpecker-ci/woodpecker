---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator changes an Account and chooses Save user
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
    api: {place: 'api::administration'}
- text: The Product saves the Account
  kind: product
  actor: administrator
  entities:
  - entity: account
    effect: changes
    facts: [Login, Email, Avatar URL, Admin]
  contexts:
    web: {place: 'web::administration::users'}
    api: {place: 'api::administration'}
---

# Edit a user

## Trigger

A person's details or admin rights must change.

## Outcome

The account has its new details and rights.

## Edge cases

- A person on the server's admin list is made an administrator again at their next sign-in; the Product warns about it.
