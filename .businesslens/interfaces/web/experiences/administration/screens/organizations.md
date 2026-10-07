---
entities:
- {entity: organization, shows: [Name]}
entryPoints:
- web: /admin/orgs
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminOrgs.vue
---

# Organizations

Every organization on the server.
