---
kind: alternative
routes:
  web: Web
steps:
- text: The Visitor chooses the forge to sign in with
  kind: actor
  actor: visitor
  entities:
  - {entity: forge, effect: reads, facts: [Forge type, URL]}
  contexts:
    web: {place: 'web::public::sign-in'}
- text: The Visitor authorizes Woodpecker on the forge
  kind: actor
  actor: visitor
  entities:
  - {entity: forge, effect: reads, facts: []}
  contexts:
    web: {place: 'web::public::sign-in'}
- text: No Account exists for that forge login and registration is open, or the login is on the admin list
  kind: condition
  entities:
  - {entity: forge, effect: reads, facts: []}
  - {entity: account, effect: reads, facts: []}
  contexts:
    web: {place: 'web::public::sign-in'}
  actor: visitor
- text: The Product creates the Account, an administrator when the login is on the admin list
  kind: product
  actor: visitor
  entities:
  - {entity: account, effect: creates, facts: [Login, Email, Avatar URL, Admin, Personal access token]}
  - {entity: administrator, effect: reads, facts: []}
  contexts:
    web: {place: 'web::public::sign-in'}
- text: The Product creates the personal Organization named after the login
  kind: product
  actor: visitor
  entities:
  - {entity: organization, effect: creates, facts: [Name, Personal]}
  contexts:
    web: {place: 'web::public::sign-in'}
- text: The Product reads the Repository permission of each enabled repository from the forge and drops access the forge no longer reports
  kind: product
  actor: visitor
  entities:
  - {entity: repository-permission, effect: changes, facts: [Pull, Push, Admin, Synced]}
  - {entity: forge, effect: reads, facts: []}
  - {entity: repository, effect: reads, facts: []}
  contexts:
    web: {place: 'web::public::sign-in'}
- text: The Product starts a session
  kind: product
  actor: visitor
  entities: []
  contexts:
    web: {place: 'web::public::sign-in'}
---

# Register on first sign-in

## Trigger

A person without an account chooses Sign in.

## Outcome

The person has an account and a personal organization and is signed in.
