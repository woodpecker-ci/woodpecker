---
availability:
- {place: 'web::signed-in'}
- {place: 'api::signed-in'}
domain: repository-actions
references:
- {kind: code, role: implementation, target: server/api/repo.go#DeleteRepo}
- {kind: code, role: implementation, target: web/src/views/repo/settings/Actions.vue}
---

# Delete repository

A repository admin deletes a repository from Woodpecker with all its pipelines, secrets and registries, after confirming that all data will be lost.
