---
appliesTo:
- type: capability
  id: trigger-pipeline
- type: entity
  id: pipeline
  effect: creates
  to: Blocked
references:
- kind: code
  role: implementation
  target: server/pipeline/gated.go
---

# Events a repository requires approval for wait for approval

Depending on the repository's approval requirements, pull requests from forks, all pull requests or all forge events start blocked, unless their author is an allowed user. Manual and cron pipelines never wait for approval.

## Rationale

Pipelines from people outside the repository could otherwise read its secrets or misuse the agents.
