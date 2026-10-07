---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: repository-actions
references:
- {kind: code, role: implementation, target: server/api/repo.go#DeleteRepo}
- {kind: code, role: implementation, target: cli/repo/repo_rm.go}
---

# Disable repository

A repository admin disables a repository: Woodpecker removes the webhook from the forge and starts no new pipelines, keeping everything else.
