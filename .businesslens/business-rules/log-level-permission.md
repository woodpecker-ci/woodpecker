---
appliesTo:
- type: entity
  id: server-settings
  effect: changes
  facts:
  - Log level
permits:
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/router/api.go
- kind: code
  role: implementation
  target: server/api/z.go#SetLogLevel
---

# Only administrators change the server's log level

Reading and setting the log level of the running server is reserved to administrators.
