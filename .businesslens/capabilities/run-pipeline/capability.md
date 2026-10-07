---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/api/pipeline.go#CreatePipeline}
- {kind: code, role: implementation, target: web/src/views/repo/RepoManualPipeline.vue}
- {kind: code, role: implementation, target: cli/pipeline/create.go}
---

# Run pipeline

A person with push access starts a pipeline by hand on the head of a branch, with an optional message and additional variables. Only workflows that run on the manual event are included.
