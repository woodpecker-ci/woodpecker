---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: forges
references:
- {kind: code, role: implementation, target: server/api/forge.go#PostForge}
- {kind: code, role: implementation, target: web/src/views/admin/forges/AdminForgeCreate.vue}
---

# Add forge

An administrator connects another forge: its type, URL, OAuth application and options. The page shows the OAuth redirect URL to register on the forge.
