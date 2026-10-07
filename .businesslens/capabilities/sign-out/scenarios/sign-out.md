---
kind: primary
routes:
  web: Web
steps:
- text: The User chooses Sign out
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in'}
- text: The Product ends the session and returns to the home page
  kind: product
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in'}
---

# Sign out

## Trigger

A signed-in person chooses Sign out.

## Outcome

The session ends and the person browses as a visitor.
