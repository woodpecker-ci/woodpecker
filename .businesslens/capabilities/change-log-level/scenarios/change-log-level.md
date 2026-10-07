---
kind: primary
routes:
  cli: CLI
  api: API
steps:
- text: The Administrator sets a log level
  kind: actor
  actor: administrator
  entities: []
  contexts:
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
- text: The Product applies the new log level to the running server
  kind: product
  actor: administrator
  entities:
  - {entity: server-settings, effect: changes, facts: [Log level]}
  contexts:
    cli: {place: 'cli::administration'}
    api: {place: 'api::administration'}
---

# Change the log level

## Trigger

An administrator needs more detail to diagnose the server, or less noise.

## Outcome

The server logs at the new level until it is changed again or the server restarts.

## Edge cases

- Without a level, the current log level is shown.
- An unknown level is refused.
