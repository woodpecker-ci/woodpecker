---
appliesTo:
- type: entity
  id: cron
  effect: removes
permits:
- related:
  - {verb: schedules, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Push, is: true}
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/router/api.go
---

# Only people with push access delete crons

A repository's crons are managed by people with push access to it, and by administrators.
