---
appliesTo:
- type: capability
  id: change-project-settings
- type: entity
  id: repository
  facts:
  - Timeout
references:
- kind: code
  role: implementation
  target: server/api/repo.go#PatchRepo
---

# Only administrators set a timeout above the server maximum

A repository admin who is not an administrator is refused a timeout above the server's maximum, and every timeout is at least one minute; a newly enabled repository gets the server's default timeout, capped at the maximum.
