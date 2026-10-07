---
appliesTo:
- type: entity
  id: pipeline
  effect: changes
  to: Canceled
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

# Only people with push access cancel a pipeline that has not started

A pending or blocked pipeline is canceled only by someone with push access to its repository, or an administrator.
