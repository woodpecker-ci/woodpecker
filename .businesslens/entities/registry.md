---
references:
- kind: code
  role: implementation
  target: server/model/registry.go#Registry
- kind: doc
  role: context
  target: docs/docs/20-usage/41-registries.md
---

# Registry

Credentials for a container registry pipelines pull images from, kept for a repository, an organization or the whole server.

## Information kept

- **Address** — the registry address
- **Username** — the user to log in as
- **Password** — the password or token, never shown again once saved
- **Read only** — set on server-wide credentials that come from the server's Docker configuration; they are shown but cannot be edited or deleted
