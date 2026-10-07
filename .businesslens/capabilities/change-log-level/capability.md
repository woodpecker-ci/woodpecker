---
availability:
- {place: 'cli::administration'}
- {place: 'api::administration'}
references:
- kind: code
  role: implementation
  target: server/api/z.go#SetLogLevel
- kind: code
  role: implementation
  target: cli/admin/loglevel/loglevel.go
---

# Change log level

An administrator reads the server's log level or sets a new one, which applies at once without restarting the server.
