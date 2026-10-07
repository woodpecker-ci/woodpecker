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
- text: The forge reports no membership in any organization the server or the forge allows
  kind: condition
  entities:
  - {entity: forge, effect: reads, facts: [Allowed organizations]}
  - {entity: organization, effect: reads, facts: []}
  contexts:
    web: {place: 'web::public::sign-in'}
  actor: visitor
- text: The Product refuses the sign-in and shows that the person may not access this organization
  kind: product
  actor: visitor
  entities:
  - {entity: organization, effect: reads, facts: []}
  contexts:
    web: {place: 'web::public::sign-in'}
---

# Refuse people outside the allowed organizations

## Trigger

A person outside the allowed organizations chooses Sign in.

## Outcome

The person stays signed out.
