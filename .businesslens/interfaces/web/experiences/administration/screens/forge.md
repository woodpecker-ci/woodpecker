---
entities:
- {entity: forge, shows: [Forge type, URL, OAuth client ID, OAuth host, Skip SSL verification, Allowed organizations, Advanced options], collects: [Forge type, URL, OAuth client ID, OAuth client secret, OAuth host, Skip SSL verification, Allowed organizations, Advanced options]}
entryPoints:
- web: /admin/forges/{forgeId}
references:
- kind: code
  role: implementation
  target: web/src/views/admin/forges/AdminForge.vue
- kind: code
  role: implementation
  target: web/src/views/admin/forges/AdminForgeCreate.vue
---

# Forge

One forge connection, created or edited.
