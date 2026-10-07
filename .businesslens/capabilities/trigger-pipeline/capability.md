---
availability:
- {place: forge-webhook}
domain: pipelines
references:
- kind: code
  role: implementation
  target: server/api/hook.go#PostHook
- kind: code
  role: implementation
  target: server/pipeline/create.go#Create
- kind: code
  role: implementation
  target: server/pipeline/gated.go
- kind: code
  role: implementation
  target: server/pipeline/cancel.go#cancelPreviousPipelines
- kind: doc
  role: context
  target: docs/docs/20-usage/20-workflow-syntax.md
---

# Trigger pipeline

The forge reports an event of an enabled repository and Woodpecker turns it into a pipeline of the workflows whose conditions match. Events from forks or other sources may wait for approval, and a new pipeline may cancel earlier active pipelines of the same branch or ref.
