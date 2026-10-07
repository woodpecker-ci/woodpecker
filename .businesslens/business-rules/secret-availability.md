---
appliesTo:
- type: entity
  id: secret
  facts:
  - Available for plugins
  - Available at events
- type: capability
  id: execute-pipeline
references:
- kind: code
  role: implementation
  target: pipeline/frontend/yaml/compiler/compiler.go#Secret.Available
---

# A secret reaches only the events and plugins it is available for

A secret limited to plugins is given only to plugin steps whose image matches, never to ordinary steps, and a secret limited to events only to pipelines of those events; all pull request events count as pull requests. A step asking for a secret it may not have fails the pipeline.
