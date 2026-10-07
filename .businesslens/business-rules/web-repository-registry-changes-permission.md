---
appliesTo:
- type: entity
  id: registry
  effect: changes
  contexts:
  - {place: 'web::signed-in::repository-registries'}
permits:
- related:
  - {verb: keeps, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Admin, is: true}
- actors: [administrator]
references:
- {kind: code, role: implementation, target: web/src/views/repo/settings/RepoSettings.vue}
---

# Only repository admins edit a repository's registries in the web UI

The web UI opens a repository's settings only for its admins and administrators, so people with push access alone edit its registries through the CLI or the API.
