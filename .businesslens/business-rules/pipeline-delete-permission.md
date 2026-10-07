---
appliesTo:
- type: entity
  id: pipeline
  effect: removes
permits:
- related:
  - {verb: has, entity: repository}
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

# Only repository admins delete a pipeline

Deleting a pipeline takes admin access to its repository, or being an administrator.
