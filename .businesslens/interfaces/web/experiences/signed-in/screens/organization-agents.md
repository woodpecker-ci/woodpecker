---
entities:
- {entity: organization, shows: [Name]}
- {entity: agent, shows: [Name, Token, Platform, Backend, Capacity, Version, Last contact, Custom labels, Filters, Disabled], collects: [Name, Filters, Disabled]}
entryPoints:
- web: /orgs/{orgId}/settings/agents
references:
- kind: code
  role: implementation
  target: web/src/views/org/settings/OrgAgents.vue
---

# Organization agents

Agents registered for a team organization; offered while the server allows user-registered agents.
