---
availability:
- {place: 'api::signed-in'}
references:
- kind: code
  role: implementation
  target: server/api/repo.go#MoveRepo
---

# Move repository

A repository admin points a Woodpecker repository at another forge repository they administer, keeping its pipelines and settings; the old name keeps leading to it.
