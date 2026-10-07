---
entities:
- {entity: secret, shows: [Name, Note, Available for plugins, Available at events], collects: [Name, Value, Note, Available for plugins, Available at events]}
entryPoints:
- web: /admin/secrets
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminSecrets.vue
---

# Global secrets

Secrets every repository on the server receives.
