---
appliesTo:
- type: entity
  id: workflow
  effect: changes
permits:
- related:
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

# Workflows change only as the Product runs them or as people with push access act on their pipeline

Approving, declining and canceling a pipeline moves its workflows and takes push access to the repository, or being an administrator; otherwise only the Product moves a workflow as it runs.
