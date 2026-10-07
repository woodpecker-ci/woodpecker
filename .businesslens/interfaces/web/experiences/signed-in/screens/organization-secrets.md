---
entities:
- {entity: organization, shows: [Name]}
- {entity: secret, shows: [Name, Note, Available for plugins, Available at events], collects: [Name, Value, Note, Available for plugins, Available at events]}
entryPoints:
- web: /orgs/{orgId}/settings/secrets
references:
- kind: code
  role: implementation
  target: web/src/views/org/settings/OrgSecrets.vue
---

# Organization secrets

A team organization's secrets, with the global secrets its repositories also receive.
