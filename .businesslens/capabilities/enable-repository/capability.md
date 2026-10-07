---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
references:
- {kind: code, role: implementation, target: server/api/repo.go#PostRepo}
- {kind: code, role: implementation, target: web/src/views/RepoAdd.vue}
- {kind: code, role: implementation, target: web/src/views/repo/settings/Actions.vue}
- {kind: code, role: implementation, target: cli/repo/repo_add.go}
---

# Enable repository

A forge admin of a repository enables it in Woodpecker: Woodpecker installs a webhook on the forge, files the repository under its owner's organization and starts with the server's default project settings. A disabled repository is enabled again the same way.
