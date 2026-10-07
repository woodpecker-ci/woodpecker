---
entities:
- {entity: account, shows: [Login, Email, Avatar URL, Admin], collects: [Login, Email, Avatar URL, Admin]}
entryPoints:
- web: /admin/users
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminUsers.vue
---

# Users

Every person with an account on the server.
