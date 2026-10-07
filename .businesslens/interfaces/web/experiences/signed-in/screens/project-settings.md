---
entities:
- {entity: repository, shows: [Pipeline path, Project visibility, Allow pull requests, Allow deployments, Approval requirements, Allowed users, Timeout, Cancel previous pipelines, Trusted, Custom trusted clone plugins], collects: [Pipeline path, Project visibility, Allow pull requests, Allow deployments, Approval requirements, Allowed users, Timeout, Cancel previous pipelines, Trusted, Custom trusted clone plugins]}
entryPoints:
- web: /repos/{repoId}/settings
references:
- kind: code
  role: implementation
  target: web/src/views/repo/settings/General.vue
---

# Project settings

A repository's project settings.
