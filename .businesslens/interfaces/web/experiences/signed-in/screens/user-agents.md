---
entities:
- {entity: agent, shows: [Name, Token, Platform, Backend, Capacity, Version, Last contact, Custom labels, Filters, Disabled], collects: [Name, Filters, Disabled]}
entryPoints:
- web: /user/agents
references:
- kind: code
  role: implementation
  target: web/src/views/user/UserAgents.vue
---

# User agents

Agents the person registered for their personal organization; offered while the server allows user-registered agents.
