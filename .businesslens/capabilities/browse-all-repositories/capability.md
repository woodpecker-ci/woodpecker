---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: repositories
references:
- {kind: code, role: implementation, target: server/api/repo.go#GetAllRepos}
- {kind: code, role: implementation, target: web/src/views/admin/AdminRepos.vue}
---

# Browse all repositories

An administrator sees every repository on the server, enabled or disabled, whoever can read it.
