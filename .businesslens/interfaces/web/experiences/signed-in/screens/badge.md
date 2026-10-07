---
entities:
- {entity: repository, shows: [Full name, Default branch]}
- {entity: workflow, shows: [Name]}
- {entity: step, shows: [Name]}
entryPoints:
- web: /repos/{repoId}/settings/badge
references:
- kind: code
  role: implementation
  target: web/src/views/repo/settings/Badge.vue
---

# Badge

Where a repository admin builds the address of the repository's status badge for a branch, events, a workflow or a step, as a URL, Markdown or HTML, with a live preview.
