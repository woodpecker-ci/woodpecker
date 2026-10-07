---
domain: project-settings
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
references:
- {kind: code, role: implementation, target: server/api/repo.go#PatchRepo}
- {kind: code, role: implementation, target: web/src/views/repo/settings/General.vue}
- {kind: code, role: implementation, target: cli/repo/repo_update.go}
---

# Change project settings

A repository admin changes how the repository's pipelines are started and run: pipeline path, project visibility, pull requests, deployments, approval requirements and allowed users, timeout, cancelling previous pipelines and custom trusted clone plugins. The CLI changes only the pipeline path, visibility, approval requirements and timeout. Only administrators raise the timeout above the server maximum.
