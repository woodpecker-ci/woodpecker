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

# Edit registry

Someone who manages registries changes a registry of a repository, a team organization, their personal organization or, as an administrator, the whole server. A password left empty keeps the current one. The web UI and the CLI never change a registry's address; the API changes the address of an organization's or the server's registry.
