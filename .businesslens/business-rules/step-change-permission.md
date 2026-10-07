---
appliesTo:
- type: entity
  id: step
  effect: changes
permits:
- related:
  - {verb: has, entity: workflow}
  - {verb: has, entity: pipeline}
  - {verb: has, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Push, is: true}
- actors:
  - administrator
- unattended: true
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#MustPush
---

# Steps change only as the Product runs them or as people with push access act on their pipeline

Approving, declining and canceling a pipeline moves its steps and takes push access to the repository, or being an administrator; otherwise only the Product moves a step as it runs.
