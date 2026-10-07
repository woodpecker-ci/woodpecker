---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator chooses Pause
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::administration::queue'}
    api: {place: 'api::administration'}
- text: The Product pauses the Queue
  kind: product
  actor: administrator
  entities:
  - entity: queue
    effect: changes
    from: Running
    to: Paused
    facts: []
  contexts:
    web: {place: 'web::administration::queue'}
    api: {place: 'api::administration'}
---

# Pause the queue

## Trigger

Agents must be updated or the server maintained.

## Outcome

The queue shows that it is paused.
