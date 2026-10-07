---
appliesTo:
- type: entity
  id: account
  effect: changes
  facts:
  - Login
  - Email
  - Avatar URL
  contexts:
  - {place: 'web::public'}
permits:
- actors:
  - visitor
references:
- kind: code
  role: implementation
  target: server/api/login.go#HandleAuth
---

# An account's forge details are refreshed only when its person signs in

At each sign-in the account's login, email and avatar are taken from the forge again.
