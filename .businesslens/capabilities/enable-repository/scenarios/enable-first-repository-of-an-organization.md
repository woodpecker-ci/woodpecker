---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User picks a forge repository to enable
  kind: actor
  actor: user
  entities:
  - entity: forge
    effect: reads
    facts: []
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    api: {place: 'api::signed-in'}
- text: The forge reports that the User administers the repository and its owner is allowed
  kind: condition
  entities:
  - entity: forge
    effect: reads
    facts: []
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    api: {place: 'api::signed-in'}
  actor: user
- text: The Product creates the Organization of the repository's owner
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: creates
    facts: [Name, Personal]
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    api: {place: 'api::signed-in'}
- text: The Product creates the Repository as Enabled with the default project settings and installs its webhook on the forge, owned by the User
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: creates
    to: Enabled
    facts: [Full name, Forge link, Default branch, Project visibility, Approval requirements, Allow pull requests, Allow deployments, Cancel previous pipelines, Timeout, Owner, Webhook]
  - entity: forge
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    api: {place: 'api::signed-in'}
- text: The Product records the User's Repository permission
  kind: product
  actor: user
  entities:
  - entity: repository-permission
    effect: creates
    facts: [Pull, Push, Admin, Synced]
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    api: {place: 'api::signed-in'}
---

# Enable the first repository of an organization

## Trigger

A person enables a repository of an owner Woodpecker has not seen yet.

## Outcome

The repository and its organization exist in Woodpecker.
