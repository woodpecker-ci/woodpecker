---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: users
references:
- {kind: code, role: implementation, target: server/api/users.go#PatchUser}
---

# Edit user

An administrator changes an account's username, email or avatar URL, and makes the person an administrator or not.
