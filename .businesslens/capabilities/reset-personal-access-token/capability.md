---
domain: cli-and-api
availability:
- {place: 'web::signed-in'}
- {place: 'api::signed-in'}
references:
- {kind: code, role: implementation, target: server/api/user.go#DeleteToken}
- {kind: code, role: implementation, target: web/src/views/user/UserCLIAndAPI.vue}
---

# Reset personal access token

A person replaces their personal access token. Every earlier token and every session of the person stops working.
