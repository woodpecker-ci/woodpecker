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

# Delete agent

Whoever may manage an agent deletes it; it can no longer connect to the server.
