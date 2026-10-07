---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User opens an Organization
  kind: actor
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::organization'}
    api: {place: 'api::signed-in'}
- text: The Product lists the Organization's enabled repositories the User may read
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::signed-in::organization'}
    api: {place: 'api::signed-in'}
---

# Browse an organization's repositories

## Trigger

A signed-in person opens an organization.

## Outcome

The person sees that organization's repositories.
