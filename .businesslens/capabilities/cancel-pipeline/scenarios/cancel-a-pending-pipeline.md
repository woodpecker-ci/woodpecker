---
kind: alternative
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Cancel on a pending Pipeline
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product skips its Workflows and cancels their Steps
  kind: product
  actor: user
  entities:
  - entity: workflow
    effect: changes
    from: Pending
    to: Skipped
    facts: []
  - entity: step
    effect: changes
    from: Pending
    to: Canceled
    facts: []
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product marks the Pipeline canceled by the User
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: changes
    from: Pending
    to: Canceled
    facts: [Cancel info]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Cancel a pipeline that has not started

## Trigger

A queued pipeline is no longer wanted.

## Outcome

The pipeline is canceled without anything running.

## Edge cases

- A pipeline that has already finished cannot be canceled.
