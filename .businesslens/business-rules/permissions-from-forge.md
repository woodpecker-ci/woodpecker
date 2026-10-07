---
appliesTo:
- type: entity
  id: repository-permission
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#SetPerm
- kind: code
  role: implementation
  target: server/api/login.go#HandleAuth
---

# Repository permissions come from the forge

What a person may do with a repository — pull, push or admin — is what the forge reports; Woodpecker reads it again when it is older than an hour, at sign-in and when the repository list is refreshed, and only for repositories on the person's own forge. Administrators have every permission on every repository.
