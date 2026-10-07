---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/pipeline/restart.go#Restart}
- {kind: code, role: implementation, target: server/api/pipeline.go#PostPipeline}
- {kind: code, role: implementation, target: cli/pipeline/start.go}
---

# Restart pipeline

A person with push access starts a pipeline that does not await approval again as a new pipeline with the same event and configuration, re-reading extension configuration, with optional extra variables.
