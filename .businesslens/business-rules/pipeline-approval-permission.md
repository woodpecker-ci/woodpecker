---
appliesTo:
- type: entity
  id: pipeline
  effect: changes
  to: Pending
permits:
- related:
  - {verb: has, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Push, is: true}
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#MustPush
---

# Only people with push access approve a pipeline

A pipeline awaiting approval is queued only when someone with push access to its repository, or an administrator, approves it.
