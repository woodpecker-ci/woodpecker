---
references:
- kind: code
  role: implementation
  target: server/model/perm.go#Perm
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#SetPerm
---

# Repository permission

What a person may do with one repository, read from the forge and kept for an hour before it is read again.

## Information kept

- **Pull** — whether the person may read the repository
- **Push** — whether the person may write to the repository
- **Admin** — whether the person administers the repository
- **Synced** — when the access was last read from the forge
