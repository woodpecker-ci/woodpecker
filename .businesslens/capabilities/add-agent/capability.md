---
availability:
- {place: 'web::administration'}
- {place: 'web::signed-in'}
- {place: 'api::administration'}
- {place: 'api::signed-in'}
references:
- {kind: code, role: implementation, target: server/api/agent.go}
- {kind: code, role: implementation, target: web/src/views/admin/AdminAgents.vue}
- {kind: code, role: implementation, target: web/src/views/org/settings/OrgAgents.vue}
- {kind: code, role: implementation, target: web/src/views/user/UserAgents.vue}
---

# Add agent

An administrator adds an agent for the whole server, or someone who administers an organization adds one for its repositories while the server allows user-registered agents. Woodpecker generates the token the agent connects with.
