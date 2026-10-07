---
entities:
- {entity: agent, shows: [Name, Token, Platform, Backend, Capacity, Version, Last contact, Custom labels, Filters, Disabled], collects: [Name, Filters, Disabled]}
entryPoints:
- web: /admin/agents
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminAgents.vue
---

# Agents

Every agent on the server, with its platform, backend, capacity, version and last contact.
