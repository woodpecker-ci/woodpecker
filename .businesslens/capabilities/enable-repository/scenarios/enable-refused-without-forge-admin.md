---
kind: validation
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
- text: The forge reports that the User does not administer the repository
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
- text: 'The Product refuses: the User has to be an admin of the repository'
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::add-repository'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Refuse enabling without forge admin rights

## Trigger

A person without admin rights on the forge tries to enable a repository.

## Outcome

Nothing is enabled.

## Edge cases

- The repository's owner is not on the server's owner allowlist: the Product refuses with "Repo owner is not allowed".
