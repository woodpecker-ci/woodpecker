---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Cancel on a running Pipeline
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
- text: The Product stops the running Workflows and their running Steps
  kind: product
  actor: user
  entities:
  - entity: workflow
    effect: changes
    from: Running
    to: Killed
    facts: []
  - entity: step
    effect: changes
    from: Running
    to: Killed
    facts: []
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product marks the Pipeline killed and canceled by the User
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: changes
    from: Running
    to: Killed
    facts: [Cancel info]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Cancel a running pipeline

## Trigger

A running pipeline is no longer wanted.

## Outcome

The pipeline is killed and shows who canceled it.
