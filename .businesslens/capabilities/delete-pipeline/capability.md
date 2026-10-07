---
availability:
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/api/pipeline.go#DeletePipeline}
- {kind: code, role: implementation, target: cli/pipeline/kill.go}
- {kind: code, role: implementation, target: cli/pipeline/purge.go}
---

# Delete pipeline

A repository admin deletes finished pipelines with their workflows, steps and logs, one at a time or by purging those of a branch or older than a limit while keeping a minimum number.
