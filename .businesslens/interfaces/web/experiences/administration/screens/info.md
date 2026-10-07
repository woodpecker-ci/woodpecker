---
entities:
- {entity: server-settings, shows: [Version]}
entryPoints:
- web: /admin
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminInfo.vue
- kind: code
  role: implementation
  target: web/src/compositions/useVersion.ts
---

# Info

The version of Woodpecker the server runs, and a notice when a newer release is available.
