---
domain: organizations
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
references:
- {kind: code, role: implementation, target: server/api/org.go#DeleteOrg}
- {kind: code, role: implementation, target: web/src/views/admin/AdminOrgs.vue}
---

# Delete organization

An administrator deletes an organization with its secrets and all repositories it owns.
