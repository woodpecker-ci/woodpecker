---
kind: primary
routes:
  cli: CLI
steps:
- text: The Visitor runs exec in a working copy, optionally with a metadata file
  kind: actor
  actor: visitor
  entities: []
  contexts:
    cli: {place: 'cli::local'}
- text: The Product runs each workflow's steps locally and prints their logs
  kind: product
  actor: visitor
  entities:
  - {entity: workflow, effect: reads, facts: []}
  - {entity: step, effect: reads, facts: []}
  contexts:
    cli: {place: 'cli::local'}
---

# Execute a pipeline locally

## Trigger

A person wants to try or debug a pipeline before pushing it.

## Outcome

The pipeline ran on their machine and its result is shown.

## Edge cases

- Secrets and variables are given on the command line; nothing is read from a server.
