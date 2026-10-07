---
entities:
- {entity: repository, shows: [Full name]}
entryPoints:
- web: /repos/{repoId}/settings/actions
references:
- kind: code
  role: implementation
  target: web/src/views/repo/settings/Actions.vue
---

# Repository actions

Where a repository is repaired, disabled, enabled again or deleted.
