---
kind: alternative
routes: {web: Web}
steps:
- text: The User opens their agents
  kind: actor
  actor: user
  entities:
  - entity: agent
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::user-agents'}
- text: The Product lists the Agents of the User's personal organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: agent
    effect: reads
    facts: [Name, Platform, Backend, Capacity, Version, Last contact, Custom labels, Filters, Disabled]
  contexts:
    web: {place: 'web::signed-in::user-agents'}
---

# Browse personal agents

## Trigger

A person checks the agents they registered.

## Outcome

The person sees their agents.
