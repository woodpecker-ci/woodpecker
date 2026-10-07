---
domain: cli-and-api
availability:
- {place: 'web::signed-in'}
- {place: 'api::signed-in'}
references:
- kind: code
  role: implementation
  target: server/api/user.go#PostToken
- kind: code
  role: implementation
  target: web/src/views/user/UserCLIAndAPI.vue
---

# View personal access token

A signed-in person sees their personal access token, with examples of using it from the CLI and the API. Viewing it changes nothing: the same token comes back until it is reset.
