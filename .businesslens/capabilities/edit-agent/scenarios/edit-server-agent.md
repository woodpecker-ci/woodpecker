---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator changes an Agent and chooses Save agent
  kind: actor
  actor: administrator
  entities:
  - entity: agent
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::administration::agents'}
    api: {place: 'api::administration'}
- text: The Product saves the Agent
  kind: product
  actor: administrator
  entities:
  - entity: agent
    effect: changes
    facts: [Name, Disabled, Filters]
  contexts:
    web: {place: 'web::administration::agents'}
    api: {place: 'api::administration'}
---

# Edit a server agent

## Trigger

An agent must be renamed, paused or narrowed.

## Outcome

The agent works under its new settings.
