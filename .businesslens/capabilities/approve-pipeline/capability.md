---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/pipeline/approve.go#Approve}
- {kind: code, role: implementation, target: cli/pipeline/approve.go}
---

# Approve pipeline

A person with push access approves a pipeline that awaits approval, so it is queued.
