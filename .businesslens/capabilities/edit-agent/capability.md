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

# Edit agent

Whoever may manage an agent renames it, disables it so it gets no new work, or sets the label filters workflows must carry to be given to it.
