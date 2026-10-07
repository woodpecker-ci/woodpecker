---
singleton: true
references:
- kind: code
  role: implementation
  target: cmd/server/flags.go
- kind: code
  role: implementation
  target: server/api/z.go
- kind: doc
  role: context
  target: docs/docs/30-administration/10-configuration/10-server.md
---

# Server settings

The settings of the one Woodpecker server, chosen by its operator when the server starts; only the log level changes while it runs.

## Information kept

- **Version** — the Woodpecker version the server runs
- **Log level** — how much the server logs; administrators change it while the server runs
- **Open registration** — whether people without an account register by signing in; closed by default
- **Admin list** — logins that are administrators and may always register
- **Allowed organizations** — organizations whose members alone may sign in, when set
- **Repository owners** — owners whose repositories may be enabled, when set
- **Default timeout** — the timeout a newly enabled repository starts with
- **Maximum timeout** — the highest timeout a repository admin who is not an administrator may set
- **User-registered agents** — Allowed unless the operator disables it: whether organizations and people register agents of their own
