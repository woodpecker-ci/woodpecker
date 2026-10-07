---
appliesTo:
- type: entity
  id: repository
  effect: removes
permits:
- related:
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Admin, is: true}
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustRepoAdmin
---

# Only repository admins delete a repository

Deleting a repository takes admin access to it, or being an administrator.
