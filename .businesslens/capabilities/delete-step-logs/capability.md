---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/api/pipeline.go#DeleteStepLogs}
- {kind: code, role: implementation, target: server/api/pipeline.go#DeletePipelineLogs}
- {kind: code, role: implementation, target: cli/pipeline/log/log_purge.go}
---

# Delete step logs

A person with push access deletes the log of one finished step, or every step log of a finished pipeline.
