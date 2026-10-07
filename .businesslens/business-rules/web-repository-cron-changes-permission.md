---
appliesTo:
- type: entity
  id: cron
  effect: changes
  contexts:
  - {place: 'web::signed-in::crons'}
permits:
- related:
  - {verb: schedules, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Admin, is: true}
- actors: [administrator]
references:
- {kind: code, role: implementation, target: web/src/views/repo/settings/RepoSettings.vue}
---

# Only repository admins edit a repository's crons in the web UI

The web UI opens a repository's settings only for its admins and administrators, so people with push access alone edit its crons through the CLI or the API.
