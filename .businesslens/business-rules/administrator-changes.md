---
appliesTo:
- type: entity
  id: account
  effect: changes
  facts:
  - Login
  - Email
  - Avatar URL
  - Admin
  contexts:
  - {place: 'web::administration'}
- type: entity
  id: forge
  effect: changes
- type: entity
  id: queue
  effect: changes
permits:
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustAdmin
---

# Only administrators edit users and forges and pause the queue

Editing accounts and forge connections and pausing or resuming the queue are reserved to administrators.
