---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User chooses Delete agent and confirms
  kind: actor
  actor: user
  entities:
  - entity: agent
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::organization-agents'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Agent
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - {entity: agent, effect: removes}
  contexts:
    web: {place: 'web::signed-in::organization-agents'}
    api: {place: 'api::signed-in'}
---

# Delete a organization agent

## Trigger

An agent is retired or its token leaked.

## Outcome

The agent can no longer connect.

## Edge cases

- An agent that is running a workflow cannot be deleted.
