---
appliesTo:
- type: capability
  id: restart-pipeline
- type: entity
  id: pipeline
references:
- kind: code
  role: implementation
  target: server/pipeline/restart.go#Restart
---

# A blocked pipeline is approved or declined, never restarted

Restarting is refused while a pipeline awaits approval.
