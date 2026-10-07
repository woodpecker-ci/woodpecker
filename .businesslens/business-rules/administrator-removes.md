---
appliesTo:
- type: entity
  id: account
  effect: removes
- type: entity
  id: organization
  effect: removes
- type: entity
  id: forge
  effect: removes
permits:
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustAdmin
---

# Only administrators delete users, organizations and forges

Deleting an account also deletes the person's personal organization and its repositories; deleting an organization deletes its secrets and repositories; deleting a forge deletes only the connection. All three are reserved to administrators.
