---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Decline on a blocked Pipeline
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
- text: The Product declines the Pipeline, its Workflows and Steps and records the User as reviewer
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: changes
    from: Blocked
    to: Declined
    facts: [Reviewer]
  - entity: workflow
    effect: changes
    from: Blocked
    to: Declined
    facts: []
  - entity: step
    effect: changes
    from: Blocked
    to: Declined
    facts: []
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Decline a pipeline

## Trigger

A maintainer does not want a pipeline from an untrusted source to run.

## Outcome

The pipeline shows that it was declined.
