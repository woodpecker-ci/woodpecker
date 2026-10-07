---
entities:
- {entity: repository, shows: [Full name]}
- {entity: registry, shows: [Address, Username], collects: [Address, Username, Password]}
entryPoints:
- web: /repos/{repoId}/settings/registries
references:
- kind: code
  role: implementation
  target: web/src/views/repo/settings/Registries.vue
---

# Repository registries

A repository's registry credentials, with the organization and global ones it also receives.
