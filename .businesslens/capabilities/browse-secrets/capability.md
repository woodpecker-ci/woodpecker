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
- {kind: code, role: implementation, target: server/api/global_secret.go}
---

# Browse secrets

Someone who manages secrets sees the secrets of a repository, a team organization, their personal organization or, as an administrator, the whole server, never their values. A repository's list also shows the organization's and the server's secrets it receives, and any signed-in person may list the server's secrets.
