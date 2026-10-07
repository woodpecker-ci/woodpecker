---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: queue
references:
- {kind: code, role: implementation, target: server/api/queue.go#GetQueueInfo}
---

# View queue

An administrator sees the workflows running on agents, pending and waiting on dependencies, with the queue's counts.
