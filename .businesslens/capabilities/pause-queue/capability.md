---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: queue
references:
- {kind: code, role: implementation, target: server/api/queue.go#PauseQueue}
---

# Pause queue

An administrator pauses the queue: running workflows finish, and no queued workflow is given to an agent.
