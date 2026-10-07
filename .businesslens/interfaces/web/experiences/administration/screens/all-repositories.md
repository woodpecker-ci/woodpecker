---
entities:
- {entity: repository, shows: [Full name]}
entryPoints:
- web: /admin/repos
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminRepos.vue
---

# All repositories

Every repository on the server, enabled or disabled.
