---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator chooses Resume
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::administration::queue'}
    api: {place: 'api::administration'}
- text: The Product resumes the Queue
  kind: product
  actor: administrator
  entities:
  - entity: queue
    effect: changes
    from: Paused
    to: Running
    facts: []
  contexts:
    web: {place: 'web::administration::queue'}
    api: {place: 'api::administration'}
---

# Resume the queue

## Trigger

Maintenance is over.

## Outcome

Queued workflows are given to agents again.
