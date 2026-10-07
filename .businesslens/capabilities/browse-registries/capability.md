---
availability:
- {place: 'web::signed-in'}
- {place: 'web::administration'}
- {place: 'cli::signed-in'}
- {place: 'cli::administration'}
- {place: 'api::signed-in'}
- {place: 'api::administration'}
references:
- {kind: code, role: implementation, target: server/api/registry.go}
- {kind: code, role: implementation, target: server/api/global_registry.go}
---

# Browse registries

Someone who manages registries sees the registry credentials of a repository, a team organization, their personal organization or, as an administrator, the whole server, never their passwords. A repository's list also shows the organization's and the server's registries it receives, and any signed-in person may list the server's registries.
