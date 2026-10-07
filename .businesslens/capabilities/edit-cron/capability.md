---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: crons
references:
- {kind: code, role: implementation, target: server/api/cron.go#PatchCron}
- {kind: code, role: implementation, target: cli/repo/cron/cron_update.go}
---

# Edit cron

A person with push access changes a cron's name, schedule, time zone, branch, variables or whether it is enabled. The CLI changes the name, schedule, branch and whether it is enabled, and enables the cron unless told otherwise.
