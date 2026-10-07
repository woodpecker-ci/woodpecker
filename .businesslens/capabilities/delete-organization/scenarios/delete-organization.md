---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Administrator chooses Delete organization and confirms
  kind: actor
  actor: administrator
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::administration::organizations'}
    api: {place: 'api::administration'}
- text: The Product deletes the Organization and its repositories
  kind: product
  actor: administrator
  entities:
  - {entity: organization, effect: removes}
  - {entity: repository, effect: removes, from: Enabled, with: organization}
  - {entity: secret, effect: removes, with: organization}
  contexts:
    web: {place: 'web::administration::organizations'}
    api: {place: 'api::administration'}
---

# Delete an organization

## Trigger

An organization is no longer built on this server.

## Outcome

The organization and its repositories are gone from Woodpecker.
