---
appliesTo:
- type: entity
  id: step
  effect: changes
  facts:
  - Log
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
  target: server/api/pipeline.go#DeleteStepLogs
---

# Only people with push access delete step logs

A step's log is written by the Product while the step runs, and deleted only by someone with push access to the repository, or an administrator.
