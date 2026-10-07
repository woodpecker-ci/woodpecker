---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator chooses Delete agent and confirms
  kind: actor
  actor: administrator
  entities:
  - entity: agent
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::administration::agents'}
    api: {place: 'api::administration'}
- text: The Product deletes the Agent
  kind: product
  actor: administrator
  entities:
  - {entity: agent, effect: removes}
  contexts:
    web: {place: 'web::administration::agents'}
    api: {place: 'api::administration'}
---

# Delete a server agent

## Trigger

An agent is retired or its token leaked.

## Outcome

The agent can no longer connect.

## Edge cases

- An agent that is running a workflow cannot be deleted.
