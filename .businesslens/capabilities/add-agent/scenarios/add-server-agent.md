---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator enters a name and chooses Save agent
  kind: actor
  actor: administrator
  entities:
  - entity: agent
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::administration::agents'}
    api: {place: 'api::administration'}
- text: The Product creates the Agent for the whole server and shows its token
  kind: product
  actor: administrator
  entities:
  - entity: agent
    effect: creates
    facts: [Name, Token, Disabled]
  contexts:
    web: {place: 'web::administration::agents'}
    api: {place: 'api::administration'}
---

# Add a server agent

## Trigger

An agent is needed for every repository on the server.

## Outcome

The agent can connect with the token shown; once connected it reports its platform, backend, capacity and version.
