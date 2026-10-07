---
appliesTo:
- type: entity
  id: account
  effect: changes
  facts:
  - Personal access token
permits:
- related:
  - {verb: uses, entity: user}
references:
- kind: code
  role: implementation
  target: server/api/user.go#DeleteToken
---

# Only the person resets their own personal access token

A personal access token is reset by the person it belongs to.
