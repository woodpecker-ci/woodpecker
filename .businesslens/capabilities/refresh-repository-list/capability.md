---
availability:
- {place: 'web::signed-in'}
- {place: 'api::signed-in'}
domain: repositories
references:
- {kind: code, role: implementation, target: server/api/user.go#RefreshRepos}
- {kind: code, role: implementation, target: web/src/views/RepoAdd.vue}
---

# Refresh repository list

A person has Woodpecker read their repositories and permissions from the forge again, so newly created forge repositories can be enabled.
