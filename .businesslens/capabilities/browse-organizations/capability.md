---
availability:
- {place: 'web::administration'}
- {place: 'cli::administration'}
- {place: 'api::administration'}
domain: organizations
references:
- {kind: code, role: implementation, target: server/api/org.go#GetOrgs}
- {kind: code, role: implementation, target: web/src/views/admin/AdminOrgs.vue}
- {kind: code, role: implementation, target: cli/admin/org/org_list.go}
---

# Browse organizations

An administrator sees every organization on the server, team and personal.
