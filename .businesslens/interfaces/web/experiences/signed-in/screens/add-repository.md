---
entities:
- {entity: repository, shows: [Full name]}
entryPoints:
- web: /repos/add
references:
- kind: code
  role: implementation
  target: web/src/views/RepoAdd.vue
---

# Add repository

The forge repositories the person may enable, with those already enabled or in conflict marked.
