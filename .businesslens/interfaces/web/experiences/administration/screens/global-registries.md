---
entities:
- {entity: registry, shows: [Address, Username], collects: [Address, Username, Password]}
entryPoints:
- web: /admin/registries
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminRegistries.vue
---

# Global registries

Registry credentials every repository on the server receives.
