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
- {kind: code, role: implementation, target: server/api/org_registry.go}
- {kind: code, role: implementation, target: server/api/global_registry.go}
- {kind: code, role: implementation, target: cli/repo/registry/registry_add.go}
---

# Add registry

Someone who manages registries adds a registry to a repository, a team organization, their personal organization or, as an administrator, the whole server.
