---
appliesTo:
- type: capability
  id: delete-pipeline
- type: capability
  id: delete-step-logs
references:
- kind: code
  role: implementation
  target: server/api/helper.go#pipelineDeleteAllowed
---

# Only finished pipelines and steps lose their logs or are deleted

Pending, running and blocked pipelines, and pending or running steps, keep their logs and cannot be deleted.
