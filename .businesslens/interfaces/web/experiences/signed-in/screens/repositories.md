---
entities:
- {entity: repository, shows: [Full name, Project visibility]}
- {entity: pipeline, shows: [Number, Event, Branch, Message]}
entryPoints:
- web: /repos
references:
- kind: code
  role: implementation
  target: web/src/views/Repos.vue
---

# Repositories

The repositories the person has access to, last visited first.
