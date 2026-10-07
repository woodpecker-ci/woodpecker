---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/pipeline/decline.go#Decline}
- {kind: code, role: implementation, target: cli/pipeline/decline.go}
---

# Decline pipeline

A person with push access declines a pipeline that awaits approval; nothing of it runs.
