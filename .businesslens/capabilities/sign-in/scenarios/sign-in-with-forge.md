---
kind: primary
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
- text: The Visitor already has an Account for that forge login
  kind: condition
  entities:
  - {entity: account, effect: reads, facts: [Login]}
  - {entity: forge, effect: reads, facts: []}
  contexts:
    web: {place: 'web::public::sign-in'}
  actor: visitor
- text: The Product refreshes the Account from the forge
  kind: product
  actor: visitor
  entities:
  - {entity: account, effect: changes, facts: [Login, Email, Avatar URL]}
  - {entity: forge, effect: reads, facts: []}
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
- text: The Product starts a session and returns the Visitor to the page they were going to
  kind: product
  actor: visitor
  entities: []
  contexts:
    web: {place: 'web::public::sign-in'}
---

# Sign in with a forge

## Trigger

A visitor chooses Sign in.

## Outcome

The person is signed in, with their repository permissions as the forge reports them.

## Edge cases

- The forge reports an OAuth error: the Product shows "Error while authenticating against OAuth provider".
- With asynchronous repository updates enabled and repositories already stored, the access refresh runs after the session starts.
- An expired or forged OAuth state is refused as invalid.
- A login on the server's admin list is made an administrator; removing it from the list does not take the role away.
