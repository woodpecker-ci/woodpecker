---
appliesTo:
- type: entity
  id: secret
  effect: reads
  facts:
  - Value
permits:
- unattended: true
references:
- kind: code
  role: implementation
  target: server/model/secret.go#Secret.Copy
---

# Only pipelines receive a secret's value

Once saved, a secret's value is never shown to anyone again; only the steps it is available to receive it.
