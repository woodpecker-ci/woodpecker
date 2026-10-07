---
kind: primary
routes:
  cli: CLI
steps:
- text: The Visitor runs lint on a configuration file or folder
  kind: actor
  actor: visitor
  entities: []
  contexts:
    cli: {place: 'cli::local'}
- text: The Product reports each error and warning with its place in the file
  kind: product
  actor: visitor
  entities: []
  contexts:
    cli: {place: 'cli::local'}
---

# Lint pipeline configuration

## Trigger

A person edits pipeline configuration.

## Outcome

They know whether the configuration is valid before pushing it.

## Edge cases

- Warnings pass unless the person asks to treat them as errors; any error fails.
