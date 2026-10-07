---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: crons
references:
- {kind: code, role: implementation, target: server/api/cron.go#PostCron}
- {kind: code, role: implementation, target: web/src/views/repo/settings/Crons.vue}
- {kind: code, role: implementation, target: cli/repo/cron/cron_add.go}
---

# Add cron

A person with push access adds a cron to a repository: a name, a schedule in a time zone, a branch and additional variables.
