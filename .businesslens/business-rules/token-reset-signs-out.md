---
appliesTo:
- type: capability
  id: reset-personal-access-token
- type: entity
  id: account
  facts:
  - Personal access token
references:
- kind: code
  role: implementation
  target: server/api/user.go#DeleteToken
---

# Resetting the personal access token signs the person out everywhere

Sessions and tokens are signed with a per-person secret that a reset replaces, so every earlier token and session stops working at once.
