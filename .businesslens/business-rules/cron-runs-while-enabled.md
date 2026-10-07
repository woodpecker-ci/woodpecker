---
appliesTo:
- type: capability
  id: run-cron
- type: entity
  id: cron
  facts:
  - Enabled
references:
- kind: code
  role: implementation
  target: server/store/datastore/cron.go
---

# A cron runs on schedule only while it and its repository are enabled

Disabled crons and crons of disabled repositories start no pipelines; Run now still starts one.
