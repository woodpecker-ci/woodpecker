---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: forges
references:
- {kind: code, role: implementation, target: server/api/forge.go#PatchForge}
- {kind: code, role: implementation, target: web/src/views/admin/forges/AdminForge.vue}
---

# Edit forge

An administrator changes a forge connection. An empty OAuth client secret keeps the current one; changes to the primary forge configured by the server's environment are reverted when the server restarts.
