---
availability:
- {place: 'web::public'}
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::public'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/api/pipeline.go#GetPipelines}
- {kind: code, role: implementation, target: web/src/views/repo/RepoPipelines.vue}
- {kind: code, role: implementation, target: cli/pipeline/list.go}
---

# Browse pipelines

A person who may read a repository sees its pipelines newest first, its branches and pull requests, and the pipelines of one branch or pull request.
