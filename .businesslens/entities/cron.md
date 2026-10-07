---
domain: crons
references:
- kind: code
  role: implementation
  target: server/model/cron.go#Cron
- kind: doc
  role: context
  target: docs/docs/20-usage/45-cron.md
---

# Cron

A schedule that starts a pipeline of a repository at set times.

## Information kept

- **Name** — unique within the repository
- **Schedule** — a cron expression
- **Timezone** — the time zone the schedule is read in; UTC by default
- **Branch** — the branch to run; the default branch when empty
- **Enabled** — whether the schedule runs
- **Variables** — additional pipeline variables for its runs
- **Next execution** — when it next runs
