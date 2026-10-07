---
entities:
- {entity: queue, shows: [Tasks, Counts]}
entryPoints:
- web: /admin/queue
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminQueue.vue
---

# Queue

The workflows running, pending and waiting on dependencies, and the queue's counts.
