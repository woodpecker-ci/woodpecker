---
kind: validation
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
- text: No Account exists for that forge login, registration is closed and the login is not on the admin list
  kind: condition
  entities:
  - {entity: forge, effect: reads, facts: []}
  - {entity: account, effect: reads, facts: []}
  contexts:
    web: {place: 'web::public::sign-in'}
  actor: visitor
- text: The Product refuses the sign-in and shows that registration is closed
  kind: product
  actor: visitor
  entities: []
  contexts:
    web: {place: 'web::public::sign-in'}
---

# Refuse registration while it is closed

## Trigger

A person without an account chooses Sign in on a server with closed registration.

## Outcome

No account is created and the person stays signed out.
