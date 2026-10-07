---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: crons
references:
- {kind: code, role: implementation, target: server/api/cron.go#DeleteCron}
- {kind: code, role: implementation, target: cli/repo/cron/cron_rm.go}
---

# Delete cron

A person with push access deletes a cron.
