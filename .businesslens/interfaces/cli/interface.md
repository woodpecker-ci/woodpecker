---
type: cli
actors:
- visitor
- user
- administrator
languages:
- en
references:
- kind: code
  role: implementation
  target: cmd/cli/app.go
- kind: doc
  role: context
  target: cli/README.md
---

# Woodpecker CLI

The `woodpecker-cli` command line tool. Signed in to a server, it works with repositories, pipelines, secrets, registries and crons, and administrators manage the server and its log level; without a server it lints and runs pipelines on the person's own machine.
