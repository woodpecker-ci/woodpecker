---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Disable repository
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::signed-in::repository-actions'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product removes the webhook from the forge and disables the Repository without an owner
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: changes
    from: Enabled
    to: Disabled
    facts: [Owner, Webhook]
  - entity: forge
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repository-actions'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product returns the User to the repository list
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repositories'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Disable a repository

## Trigger

A repository admin no longer wants Woodpecker to build a repository.

## Outcome

The repository is disabled; enabling it again restores it.

## Edge cases

- The repository no longer exists on the forge: it is disabled anyway.
