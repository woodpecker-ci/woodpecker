---
availability:
- {place: 'cli::local'}
references:
- kind: code
  role: implementation
  target: cli/lint/lint.go
- kind: code
  role: implementation
  target: pipeline/frontend/yaml/linter/linter.go
- kind: doc
  role: context
  target: docs/docs/20-usage/72-linter.md
---

# Lint pipeline configuration

Anyone checks pipeline configuration files on their own machine against the pipeline syntax and the linter's rules, without a server.
