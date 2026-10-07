---
type: web
actors:
- visitor
- user
- administrator
entryPoints:
- web: /
references:
- kind: code
  role: implementation
  target: web/src/router.ts
- kind: code
  role: implementation
  target: server/router/router.go#Load
---

# Web UI

The Woodpecker web application served by the server, where people sign in, follow pipelines and manage repositories, organizations and the server.
