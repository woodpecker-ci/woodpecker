---
entities:
- {entity: repository, shows: [Config extension endpoint, Config extension exclusive, Registry extension endpoint, Secret extension endpoint, Include netrc credentials], collects: [Config extension endpoint, Config extension exclusive, Registry extension endpoint, Secret extension endpoint, Include netrc credentials]}
entryPoints:
- web: /repos/{repoId}/settings/extensions
references:
- kind: code
  role: implementation
  target: web/src/views/repo/settings/Extensions.vue
---

# Extensions

The HTTP extensions a repository asks for configuration, registries and secrets.
