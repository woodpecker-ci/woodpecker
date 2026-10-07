---
appliesTo:
- type: entity
  id: pipeline
  effect: creates
permits:
- related:
  - {verb: has, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Push, is: true}
- actors:
  - administrator
- actors:
  - forge
- unattended: true
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#MustPush
- kind: code
  role: implementation
  target: server/router/api.go
---

# Pipelines start only from forge events, schedules or people with push access

The forge starts pipelines with its events and crons start them on schedule; by hand, people with push access run, restart and deploy pipelines, as do administrators.
