---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Approve on a blocked Pipeline
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number, Author, Changed files]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product queues the Pipeline with its Workflows and Steps and records the User as reviewer
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: changes
    from: Blocked
    to: Pending
    facts: [Reviewer]
  - entity: workflow
    effect: changes
    from: Blocked
    to: Pending
    facts: []
  - entity: step
    effect: changes
    from: Blocked
    to: Pending
    facts: []
  - entity: queue
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Approve a pipeline

## Trigger

A maintainer has checked a pipeline from an untrusted source.

## Outcome

The pipeline is queued and runs.

## Edge cases

- A pipeline that is not blocked cannot be approved.
- When the configuration no longer parses, the pipeline ends in Error.
