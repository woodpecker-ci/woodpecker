---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Visitor chooses Download on a Step's log
  kind: actor
  actor: visitor
  entities:
  - entity: step
    effect: reads
    facts: [Name, Log]
  contexts:
    web: {place: 'web::public::pipeline'}
    api: {place: 'api::public'}
---

# Download a public step log

## Trigger

Someone not signed in needs a public step's whole log.

## Outcome

They have the log as a file.
