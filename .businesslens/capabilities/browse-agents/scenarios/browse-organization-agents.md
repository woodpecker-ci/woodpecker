---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User opens the agents of an Organization
  kind: actor
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: agent
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::organization-agents'}
    api: {place: 'api::signed-in'}
- text: The Product lists the Agents registered for the Organization
  kind: product
  actor: user
  entities:
  - entity: agent
    effect: reads
    facts: [Name, Platform, Backend, Capacity, Version, Last contact, Custom labels, Filters, Disabled]
  - entity: organization
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::organization-agents'}
    api: {place: 'api::signed-in'}
---

# Browse an organization's agents

## Trigger

An organization admin checks the organization's agents.

## Outcome

The admin sees the organization's agents.
