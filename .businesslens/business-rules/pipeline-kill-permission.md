---
appliesTo:
- type: entity
  id: pipeline
  effect: changes
  to: Killed
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
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#MustPush
---

# Only people with push access, or a newer forge event, stop a running pipeline

People with push access and administrators cancel a running pipeline; a newer event of the same branch or ref supersedes it when the repository cancels previous pipelines.
