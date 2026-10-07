---
references:
- kind: code
  role: implementation
  target: server/model/perm.go#OrgPerm
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustOrgMember
---

# Organization membership

A person's membership of a team organization, read from the forge.

## Information kept

- **Role** — Member or Admin, as the forge reports it
