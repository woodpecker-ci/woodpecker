---
availability:
- {place: 'web::administration'}
- {place: 'web::signed-in'}
- {place: 'api::administration'}
- {place: 'api::signed-in'}
references:
- {kind: code, role: implementation, target: server/api/agent.go}
- {kind: code, role: implementation, target: web/src/components/agent/AgentManager.vue}
---

# Browse agents

Whoever may manage agents sees them: administrators every agent of the server, an organization's admins the agents registered for it, and a person the agents of their personal organization, each with its platform, backend, capacity, version, labels, filters and last contact.
