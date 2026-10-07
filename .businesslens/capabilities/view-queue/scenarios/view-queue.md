---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator opens the Queue
  kind: actor
  actor: administrator
  entities:
  - entity: queue
    effect: reads
    facts: [Tasks, Counts]
  contexts:
    web: {place: 'web::administration::queue'}
    api: {place: 'api::administration'}
---

# View the queue

## Trigger

An administrator wants to know why pipelines wait.

## Outcome

They see what runs where and what waits.
