---
domain: repositories
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
references:
- {kind: code, role: implementation, target: server/api/repo.go#RepairAllRepos}
- {kind: code, role: implementation, target: web/src/views/admin/AdminRepos.vue}
---

# Repair all repositories

An administrator repairs every enabled repository on the server at once.
