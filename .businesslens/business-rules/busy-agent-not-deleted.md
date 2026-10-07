---
appliesTo:
- type: capability
  id: delete-agent
- type: entity
  id: agent
references:
- kind: code
  role: implementation
  target: server/api/agent.go
---

# An agent running workflows cannot be deleted

Deleting an agent is refused while it runs a workflow; a deleted agent can no longer connect.
