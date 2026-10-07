---
appliesTo:
- type: entity
  id: workflow
  facts:
  - Depends on
- type: capability
  id: execute-pipeline
references:
- kind: code
  role: implementation
  target: server/model/task.go
- kind: code
  role: implementation
  target: server/queue/fifo.go
- kind: doc
  role: context
  target: docs/docs/20-usage/25-workflows.md
---

# A workflow waits for the workflows it depends on and for a free slot of its concurrency limit

A workflow is given to an agent only after every workflow it depends on has finished as its status conditions require, by default successfully; otherwise it is skipped. A workflow with a concurrency limit stays queued while that many of it already run, and free slots go to the earliest pipeline first; nothing is canceled for it.
