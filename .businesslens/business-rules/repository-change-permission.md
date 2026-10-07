---
appliesTo:
- type: entity
  id: repository
  effect: changes
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
- kind: code
  role: implementation
  target: server/api/repo.go#PostRepo
---

# Only repository admins change a repository

Project settings, extensions, repairing, disabling, enabling again and assuming ownership take admin access to the repository, or being an administrator.
