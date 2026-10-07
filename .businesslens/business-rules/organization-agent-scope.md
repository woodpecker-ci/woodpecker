---
appliesTo:
- type: entity
  id: agent
- type: capability
  id: execute-pipeline
references:
- kind: code
  role: implementation
  target: server/model/agent.go#Agent.CanAccessRepo
---

# An organization's agent runs only that organization's workflows

Server agents run workflows of every repository; an agent registered for an organization runs only workflows of that organization's repositories. A disabled agent receives no new workflows.
