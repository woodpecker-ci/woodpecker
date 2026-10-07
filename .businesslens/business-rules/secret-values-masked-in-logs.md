---
appliesTo:
- type: entity
  id: step
  facts:
  - Log
- type: entity
  id: secret
  facts:
  - Value
- type: capability
  id: execute-local-pipeline
references:
- kind: code
  role: implementation
  target: agent/logger.go
- kind: code
  role: implementation
  target: pipeline/shared/secrets_writer.go
- kind: code
  role: implementation
  target: cli/exec/exec.go
- kind: doc
  role: context
  target: docs/docs/20-usage/40-secrets.md
---

# Secret values are masked in step logs

Every value of a secret a workflow receives is replaced with asterisks in its steps' logs before they are stored or printed, including when a pipeline runs locally. Values of three characters or fewer, and secrets a step fetches from elsewhere, are not masked.
