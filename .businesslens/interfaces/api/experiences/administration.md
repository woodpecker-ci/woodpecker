---
actors:
- administrator
access: restricted
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustAdmin
---

# Administration

Calls accepted only from administrators: users, organizations, all repositories, global secrets and registries, agents, the queue, forges and the server's log level.
