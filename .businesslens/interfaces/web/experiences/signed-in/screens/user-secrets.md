---
entities:
- {entity: secret, shows: [Name, Note, Available for plugins, Available at events], collects: [Name, Value, Note, Available for plugins, Available at events]}
entryPoints:
- web: /user/secrets
references:
- kind: code
  role: implementation
  target: web/src/views/user/UserSecrets.vue
---

# User secrets

The secrets of the person's personal organization.
