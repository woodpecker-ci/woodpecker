---
availability:
- {place: 'web::administration'}
- {place: 'cli::administration'}
- {place: 'api::administration'}
domain: users
references:
- {kind: code, role: implementation, target: server/api/users.go#PostUser}
- {kind: code, role: implementation, target: web/src/views/admin/AdminUsers.vue}
- {kind: code, role: implementation, target: cli/admin/user/user_add.go}
---

# Add user

An administrator adds an account for a forge login before the person signs in.
