---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User chooses Download on a Step's log
  kind: actor
  actor: user
  entities:
  - entity: step
    effect: reads
    facts: [Name, Log]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    api: {place: 'api::signed-in'}
---

# Download a step log

## Trigger

A signed-in person needs a step's whole log.

## Outcome

They have the log as a file.
