---
kind: primary
routes: {web: Web, cli: CLI, api: API}
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
    cli: {place: 'cli::signed-in'}
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
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
  actor: user
- text: The Organization of the repository's owner already exists
  kind: condition
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
  actor: user
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
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product records the User's Repository permission
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: repository-permission
    effect: creates
    facts: [Pull, Push, Admin, Synced]
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product opens the new Repository
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::signed-in::repository'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Enable a forge repository

## Trigger

A person wants Woodpecker to build a repository.

## Outcome

The repository is enabled, its webhook is on the forge and the person lands on its activity.

## Edge cases

- The repository is already enabled: the Product refuses with a conflict.
- The repository was recreated on the forge while an outdated entry with the same name exists: the Product refuses with a conflict until the stale entry is removed.
