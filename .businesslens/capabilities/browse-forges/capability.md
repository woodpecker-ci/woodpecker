---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: forges
references:
- {kind: code, role: implementation, target: server/api/forge.go#GetForges}
- {kind: code, role: implementation, target: web/src/views/admin/forges/AdminForges.vue}
---

# Browse forges

An administrator sees the forges the server is connected to with their settings, never the OAuth client secret. Without being an administrator, anyone sees only each forge's type and URL, which the sign-in page offers.
