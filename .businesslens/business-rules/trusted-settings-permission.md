---
appliesTo:
- type: entity
  id: repository
  effect: changes
  facts:
  - Trusted
permits:
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/api/repo.go#PatchRepo
---

# Only administrators change a repository's trusted settings

The trusted network, volumes and security settings let pipelines escape their sandbox, so only administrators change them.
