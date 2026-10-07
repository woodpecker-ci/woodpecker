---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User selects a branch, adds a message and variables, and chooses Run pipeline
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
- text: The Product reads the pipeline configuration of the Repository at that commit
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Pipeline path, Config extension endpoint, Config extension exclusive]
  - entity: pipeline
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::manual-pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product creates a Pipeline in Pending with a Workflow for each configuration file and the Steps it defines
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: creates
    to: Pending
    facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Additional pipeline variables]
  - entity: workflow
    effect: creates
    to: Pending
    facts: [Name, Platform, Matrix variables, Depends on]
    with: pipeline
  - entity: step
    effect: creates
    to: Pending
    facts: [Name, Type]
    with: workflow
  contexts:
    web: {place: 'web::signed-in::manual-pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product opens the new Pipeline
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Run a manual pipeline

## Trigger

A person wants a pipeline without pushing a commit.

## Outcome

A manual pipeline is queued and the person follows it.
