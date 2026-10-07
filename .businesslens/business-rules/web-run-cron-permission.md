---
appliesTo:
- type: entity
  id: pipeline
  effect: creates
  contexts:
  - {place: 'web::signed-in::crons'}
permits:
- related:
  - {verb: has, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Admin, is: true}
- actors: [administrator]
references:
- {kind: code, role: implementation, target: web/src/views/repo/settings/RepoSettings.vue}
---

# Only repository admins run a cron now in the web UI

The crons of a repository sit in its settings, which the web UI opens only for its admins and administrators; people with push access alone run a cron now through the API.
