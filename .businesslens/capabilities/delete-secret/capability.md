---
availability:
- {place: 'web::signed-in'}
- {place: 'web::administration'}
- {place: 'cli::signed-in'}
- {place: 'cli::administration'}
- {place: 'api::signed-in'}
- {place: 'api::administration'}
references:
- {kind: code, role: implementation, target: server/api/repo_secret.go}
- {kind: code, role: implementation, target: server/api/org_secret.go}
- {kind: code, role: implementation, target: server/api/global_secret.go}
- {kind: code, role: implementation, target: cli/repo/secret/secret_add.go}
---

# Delete secret

Someone who manages secrets deletes a secret of a repository, a team organization, their personal organization or, as an administrator, the whole server.
