---
appliesTo:
- type: entity
  id: forge
  effect: reads
  facts:
  - OAuth client secret
permits: []
references:
- kind: code
  role: implementation
  target: server/model/forge.go#Forge
- kind: code
  role: implementation
  target: server/api/forge.go#GetForges
---

# Nobody reads a forge's OAuth client secret

Once saved, a forge's OAuth client secret is never shown, not even to administrators; leaving it empty when editing keeps it. People who are not administrators see only each forge's type and URL.
