---
appliesTo:
- type: entity
  id: account
  effect: creates
  contexts:
  - {place: 'web::administration'}
  - {place: 'cli::administration'}
- type: entity
  id: forge
  effect: creates
permits:
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustAdmin
---

# Only administrators add users and forges

Administrators add accounts in the administration area and connect forges; people otherwise get an account by signing in.
