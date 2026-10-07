---
kind: person
acts: external
relations:
- entity: account
  verb: uses
  cardinality: one-to-one
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustAdmin
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#SetPerm
---

# Administrator

A signed-in person who administers the Woodpecker server: its users, organizations, repositories, global secrets and registries, agents, queue and forges. Administrators also have full access to every repository and organization.
