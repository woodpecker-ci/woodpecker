---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/pipeline/cancel.go#Cancel}
- {kind: code, role: implementation, target: server/api/pipeline.go#CancelPipeline}
- {kind: code, role: implementation, target: cli/pipeline/stop.go}
---

# Cancel pipeline

A person with push access cancels a pipeline that is pending, running or awaiting approval: running workflows are stopped, workflows that have not started are skipped and their steps canceled. A pipeline whose workflows were all pending ends Canceled; any other ends Killed. The web UI offers Cancel only on pending and running pipelines.
