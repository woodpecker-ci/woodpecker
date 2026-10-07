---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User chooses Refresh repository list
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    api: {place: 'api::signed-in'}
- text: The Product reads the User's repositories and Repository permission from the forge again
  kind: product
  actor: user
  entities:
  - entity: repository-permission
    effect: changes
    facts: [Pull, Push, Admin, Synced]
  - entity: forge
    effect: reads
    facts: []
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    api: {place: 'api::signed-in'}
---

# Refresh the repository list

## Trigger

A person misses a repository they just created on the forge.

## Outcome

The list of forge repositories and the person's permissions are current.
