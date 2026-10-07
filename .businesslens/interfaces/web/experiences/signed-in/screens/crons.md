---
entities:
- {entity: repository, shows: [Full name]}
- {entity: cron, shows: [Name, Schedule, Timezone, Branch, Enabled, Variables, Next execution], collects: [Name, Schedule, Timezone, Branch, Enabled, Variables]}
entryPoints:
- web: /repos/{repoId}/settings/crons
references:
- kind: code
  role: implementation
  target: web/src/views/repo/settings/Crons.vue
---

# Crons

A repository's crons and when each runs next.
