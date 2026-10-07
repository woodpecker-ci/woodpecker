---
domain: project-settings
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
references:
- kind: code
  role: implementation
  target: server/api/repo.go#PatchRepo
- kind: code
  role: implementation
  target: web/src/views/repo/settings/General.vue
- kind: code
  role: implementation
  target: cli/repo/repo_update.go
---

# Change repository trust

An administrator turns a repository's trusted network, volumes and security settings on or off, letting its pipelines use those privileged options. The settings show among the project settings only to administrators.
