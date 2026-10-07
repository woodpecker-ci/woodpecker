---
kind: primary
routes: {api: API}
steps:
- text: The User names the forge repository to move the Repository to
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: forge
    effect: reads
    facts: []
  contexts:
    api: {place: 'api::signed-in'}
- text: The forge reports that the User administers the target repository
  kind: condition
  actor: user
  entities:
  - entity: forge
    effect: reads
    facts: []
  - entity: repository
    effect: reads
    facts: []
  contexts:
    api: {place: 'api::signed-in'}
- text: The Product points the Repository at the target, keeps the old name leading to it and installs its webhook again
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: changes
    facts: [Full name, Forge link, Default branch, Webhook]
  - entity: forge
    effect: reads
    facts: []
  contexts:
    api: {place: 'api::signed-in'}
---

# Move a repository

## Trigger

A repository moved to another owner or name on the forge.

## Outcome

The repository's history and settings continue under the new forge repository.

## Edge cases

- Without admin rights on the target repository, the move is refused.
