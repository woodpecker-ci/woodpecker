---
kind: alternative
routes:
  api: API
steps:
- text: The Visitor requests the server version
  kind: actor
  actor: visitor
  entities: []
  contexts:
    api: {place: 'api::public'}
- text: The Product answers with the version the server runs
  kind: product
  actor: visitor
  entities:
  - {entity: server-settings, effect: reads, facts: [Version]}
  contexts:
    api: {place: 'api::public'}
---

# Request the server version

## Trigger

A tool needs to know which Woodpecker version a server runs.

## Outcome

The caller has the version.
