---
appliesTo:
- type: entity
  id: secret
- type: capability
  id: execute-pipeline
references:
- kind: code
  role: implementation
  target: server/services/secret/db.go
---

# A repository secret overrides an organization secret of the same name, which overrides a global one

When secrets of several levels share a name, a pipeline receives only the one closest to its repository.
