---
availability:
- {place: 'web::public'}
- {place: 'web::signed-in'}
- {place: 'api::public'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/api/pipeline.go#DownloadStepLogs}
---

# Download step log

A person who may read a repository downloads a step's complete log.
