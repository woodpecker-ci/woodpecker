---
kind: edge
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User selects a branch and chooses Run pipeline
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::manual-pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: No workflow of the Repository's configuration runs on the manual event
  kind: condition
  entities:
  - entity: repository
    effect: reads
    facts: [Pipeline path]
  - entity: workflow
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::manual-pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
  actor: user
- text: The Product starts nothing and warns that no matching workflows were found
  kind: product
  actor: user
  entities:
  - entity: workflow
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repository'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Run with no manual workflows

## Trigger

A person runs a pipeline whose configuration has no manual workflow.

## Outcome

No pipeline is created.
