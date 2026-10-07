---
entities:
- {entity: organization, shows: [Name]}
- {entity: repository, shows: [Full name]}
entryPoints:
- web: /orgs/{orgId}
references:
- kind: code
  role: implementation
  target: web/src/views/org/OrgRepos.vue
---

# Organization

An organization's enabled repositories.
