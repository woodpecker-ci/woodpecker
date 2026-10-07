---
availability:
- {place: 'web::signed-in'}
- {place: 'api::signed-in'}
domain: crons
references:
- {kind: code, role: implementation, target: server/cron/cron.go#Run}
- {kind: code, role: implementation, target: server/cron/cron.go#CreatePipeline}
- {kind: code, role: implementation, target: server/api/cron.go#RunCron}
---

# Run cron

A cron starts a pipeline with the cron event on the head of its branch, with its variables: at each scheduled time while it and its repository are enabled, or at once when someone with push access chooses Run now, even while the cron is disabled.
