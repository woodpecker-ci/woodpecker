---
appliesTo:
- type: entity
  id: registry
- type: capability
  id: execute-pipeline
references:
- kind: code
  role: implementation
  target: server/services/registry/db.go
- kind: code
  role: implementation
  target: server/services/registry/combined.go
- kind: code
  role: implementation
  target: pipeline/frontend/yaml/utils/image.go
- kind: doc
  role: context
  target: docs/docs/20-usage/41-registries.md
---

# Repository registries override organization ones of the same address, then global ones

A step's image is pulled with the credentials whose address matches the image's registry host. When several levels hold the same address, the one closest to the repository wins, and credentials saved in Woodpecker win over read-only ones from the server's Docker configuration. Registry credentials are never given to the steps themselves.
