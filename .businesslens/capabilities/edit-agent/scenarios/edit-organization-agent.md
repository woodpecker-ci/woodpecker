---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User changes an Agent and chooses Save agent
  kind: actor
  actor: user
  entities:
  - entity: agent
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::organization-agents'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Agent
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: agent
    effect: changes
    facts: [Name, Disabled, Filters]
  contexts:
    web: {place: 'web::signed-in::organization-agents'}
    api: {place: 'api::signed-in'}
---

# Edit a organization agent

## Trigger

An agent must be renamed, paused or narrowed.

## Outcome

The agent works under its new settings.
