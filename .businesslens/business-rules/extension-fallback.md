---
appliesTo:
- type: entity
  id: repository
  facts:
  - Config extension endpoint
  - Registry extension endpoint
  - Secret extension endpoint
  - Include netrc credentials
- type: capability
  id: trigger-pipeline
- type: capability
  id: execute-pipeline
references:
- kind: code
  role: implementation
  target: server/services/secret/combined.go
- kind: code
  role: implementation
  target: server/services/registry/with_extension.go
- kind: code
  role: implementation
  target: server/services/config/http.go
- kind: code
  role: implementation
  target: server/pipeline/create.go
- kind: code
  role: implementation
  target: server/pipeline/restart.go
- kind: doc
  role: context
  target: docs/docs/20-usage/72-extensions/index.md
---

# Extensions take precedence, and a failing secret or registry extension falls back to what Woodpecker keeps

Secrets and registries an extension returns override those Woodpecker keeps under the same name or address; when the extension fails or returns nothing, pipelines receive the kept ones. A configuration extension that answers with no content keeps the configuration Woodpecker found; when it fails, a new pipeline falls back to that configuration or ends in error without one, and a restart is refused. Forge credentials are sent to an extension only while Include netrc credentials is on for it.
