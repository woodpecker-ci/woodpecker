---
availability:
- {place: 'web::administration'}
- {place: 'cli::administration'}
- {place: 'api::administration'}
domain: users
references:
- {kind: code, role: implementation, target: server/api/users.go#GetUsers}
- {kind: code, role: implementation, target: web/src/views/admin/AdminUsers.vue}
- {kind: code, role: implementation, target: cli/admin/user/user_list.go}
---

# Browse users

An administrator sees every account on the server, and which are administrators.
