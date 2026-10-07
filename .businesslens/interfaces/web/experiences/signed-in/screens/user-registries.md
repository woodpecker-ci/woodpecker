---
entities:
- {entity: registry, shows: [Address, Username], collects: [Address, Username, Password]}
entryPoints:
- web: /user/registries
references:
- kind: code
  role: implementation
  target: web/src/views/user/UserRegistries.vue
---

# User registries

The registry credentials of the person's personal organization.
