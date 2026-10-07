---
references:
- kind: code
  role: implementation
  target: server/model/secret.go#Secret
- kind: doc
  role: context
  target: docs/docs/20-usage/40-secrets.md
---

# Secret

A named value pipelines receive as an environment variable, kept for a repository, an organization (including a person's personal organization) or the whole server.

## Information kept

- **Name** — the name pipelines refer to
- **Value** — the secret value, never shown again once saved
- **Note** — a description for the people who manage it
- **Available for plugins** — when set, only these plugin images receive the secret
- **Available at events** — the pipeline events it is given to; empty means every event
