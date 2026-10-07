---
kind: primary
routes: {cli: CLI, api: API}
steps:
- text: The User runs chown for a Repository
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product makes the User the Repository's owner
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: changes
    facts: [Owner]
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Assume ownership of a repository

## Trigger

The current owner left, or their forge credentials stopped working.

## Outcome

Woodpecker acts on the forge with the new owner's credentials.
