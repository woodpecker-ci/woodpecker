---
references:
- kind: code
  role: implementation
  target: server/model/agent.go#Agent
---

# Agent

A worker that connects to the server with a token and runs workflows. It serves the whole server, or only one organization's repositories.

## Information kept

- **Name** — shown to administrators; the agent's hostname when not given
- **Token** — the secret the agent connects with
- **Platform** — the platform it reports
- **Backend** — docker, kubernetes, local or a custom backend
- **Capacity** — how many workflows it runs at once
- **Version** — the Woodpecker version it runs
- **Last contact** — when it last reached the server
- **Custom labels** — labels it offers for matching workflows
- **Filters** — labels workflows must carry to be given to it
- **Disabled** — set by Disable agent: the agent takes no new tasks
