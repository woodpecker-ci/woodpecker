---
entities:
- {entity: repository, shows: [Full name]}
- {entity: secret, shows: [Name, Note, Available for plugins, Available at events], collects: [Name, Value, Note, Available for plugins, Available at events]}
entryPoints:
- web: /repos/{repoId}/settings/secrets
references:
- kind: code
  role: implementation
  target: web/src/views/repo/settings/Secrets.vue
---

# Repository secrets

A repository's secrets, with the organization and global secrets it also receives.
