---
availability:
- {place: 'web::administration'}
- {place: 'cli::administration'}
- {place: 'api::administration'}
domain: users
references:
- {kind: code, role: implementation, target: server/api/users.go#DeleteUser}
- {kind: code, role: implementation, target: cli/admin/user/user_rm.go}
---

# Delete user

An administrator deletes an account, and with it the repositories the person owns.
