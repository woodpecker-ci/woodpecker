---
availability:
- {place: 'web::public'}
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::public'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/api/pipeline.go#GetPipeline}
- {kind: code, role: implementation, target: server/api/stream.go#LogStreamSSE}
- {kind: code, role: implementation, target: web/src/views/repo/pipeline/Pipeline.vue}
- {kind: code, role: implementation, target: cli/pipeline/show.go}
- {kind: code, role: implementation, target: cli/pipeline/log/log_show.go}
---

# View pipeline

A person who may read a repository follows one pipeline: its workflows and steps, each step's log as it is written, the configuration it ran, its changed files and its errors and warnings.
