---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: queue
references:
- {kind: code, role: implementation, target: server/api/queue.go#ResumeQueue}
---

# Resume queue

An administrator resumes a paused queue.
