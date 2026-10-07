---
availability:
- {place: 'web::administration'}
- {place: 'api::administration'}
domain: forges
references:
- {kind: code, role: implementation, target: server/api/forge.go#DeleteForge}
---

# Delete forge

An administrator deletes a forge connection. Its repositories and accounts stay in Woodpecker, but nobody signs in through it any more.
