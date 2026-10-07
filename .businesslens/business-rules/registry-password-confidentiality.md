---
appliesTo:
- type: entity
  id: registry
  effect: reads
  facts:
  - Password
permits:
- unattended: true
references:
- kind: code
  role: implementation
  target: server/model/registry.go#Registry.Copy
---

# Only pipelines receive a registry password

Once saved, a registry password is never shown to anyone again; only pipelines pulling images receive it.
