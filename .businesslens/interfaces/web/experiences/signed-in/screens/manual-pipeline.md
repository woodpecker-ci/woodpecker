---
entities:
- {entity: pipeline, collects: [Branch, Message, Additional pipeline variables]}
entryPoints:
- web: /repos/{repoId}/manual
references:
- kind: code
  role: implementation
  target: web/src/views/repo/RepoManualPipeline.vue
---

# Manual pipeline

Where a person starts a pipeline by hand on a branch they choose.
