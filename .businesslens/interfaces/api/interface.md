---
type: api
actors:
- visitor
- user
- administrator
entryPoints:
- api: /api
languages:
- en
references:
- kind: code
  role: implementation
  target: server/router/api.go
- kind: code
  role: implementation
  target: cmd/server/openapi.go
- kind: code
  role: implementation
  target: web/src/views/user/UserCLIAndAPI.vue
---

# API

The Woodpecker REST API the server publishes with its Swagger UI and the docs site's API page. People call it with their personal access token, as the web UI's CLI & API page shows; anyone reads public repositories, status badges and the server version without one.
