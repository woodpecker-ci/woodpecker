---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator opens the agents
  kind: actor
  actor: administrator
  entities:
  - entity: agent
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::administration::agents'}
    api: {place: 'api::administration'}
- text: The Product lists every Agent of the server
  kind: product
  actor: administrator
  entities:
  - entity: agent
    effect: reads
    facts: [Name, Platform, Backend, Capacity, Version, Last contact, Custom labels, Filters, Disabled]
  contexts:
    web: {place: 'web::administration::agents'}
    api: {place: 'api::administration'}
---

# Browse the server's agents

## Trigger

An administrator checks which agents serve the server.

## Outcome

The administrator sees each agent and when it last made contact.
