---
entities:
- {entity: repository, shows: [Full name, Forge link]}
- {entity: pipeline, shows: [Number, Event, Branch, Message, Author, Created, Duration]}
entryPoints:
- web: /repos/{repoId}
references:
- kind: code
  role: implementation
  target: web/src/views/repo/RepoPipelines.vue
---

# Repository

A public repository's activity: its pipelines, newest first, and its branches and pull requests with their pipelines.
