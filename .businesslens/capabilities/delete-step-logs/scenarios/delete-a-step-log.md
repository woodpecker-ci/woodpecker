---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Delete on a finished Step's log and confirms
  kind: actor
  actor: user
  entities:
  - entity: step
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Step's log
  kind: product
  actor: user
  entities:
  - entity: step
    effect: changes
    facts: [Log]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Delete a step log

## Trigger

A log holds something that must not stay.

## Outcome

The step shows no logs.

## Edge cases

- A pending or running step's log cannot be deleted.
