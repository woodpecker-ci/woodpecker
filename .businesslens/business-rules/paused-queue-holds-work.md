---
appliesTo:
- type: entity
  id: queue
- type: capability
  id: execute-pipeline
references:
- kind: code
  role: implementation
  target: server/queue/fifo.go
---

# A paused queue gives no workflow to an agent

While the queue is paused, running workflows finish and queued ones wait until it is resumed.
