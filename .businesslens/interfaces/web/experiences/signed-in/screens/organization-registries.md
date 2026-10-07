---
entities:
- {entity: organization, shows: [Name]}
- {entity: registry, shows: [Address, Username], collects: [Address, Username, Password]}
entryPoints:
- web: /orgs/{orgId}/settings/registries
references:
- kind: code
  role: implementation
  target: web/src/views/org/settings/OrgRegistries.vue
---

# Organization registries

A team organization's registry credentials.
