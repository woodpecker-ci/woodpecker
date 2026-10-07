---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User enters a name and chooses Save agent
  kind: actor
  actor: user
  entities:
  - entity: agent
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::user-agents'}
    api: {place: 'api::signed-in'}
- text: The Product creates the Agent for the User's personal organization and shows its token
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: agent
    effect: creates
    facts: [Name, Token, Disabled]
  contexts:
    web: {place: 'web::signed-in::user-agents'}
    api: {place: 'api::signed-in'}
---

# Add a personal agent

## Trigger

An agent is needed for the person's own repositories.

## Outcome

The agent can connect with the token shown; once connected it reports its platform, backend, capacity and version.
