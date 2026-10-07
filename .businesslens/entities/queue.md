---
domain: queue
references:
- kind: code
  role: implementation
  target: server/queue/queue.go
- kind: code
  role: implementation
  target: server/api/queue.go
---

# Queue

The server's queue of workflows waiting for agents.

## Information kept

- **Tasks** — the running, pending and waiting workflows, with the agent each runs on
- **Counts** — free workers, running, pending, waiting on dependencies and completed tasks

## States

### Running

Agents receive queued workflows.

### Paused

No queued workflow is given to an agent until the queue is resumed.
