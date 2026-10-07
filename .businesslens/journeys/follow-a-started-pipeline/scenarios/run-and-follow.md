---
kind: primary
result: achieved
routes:
  web: Web
steps:
- text: The User selects a branch and chooses Run pipeline
  kind: actor
  actor: user
  capability: run-pipeline
  entities:
  - {entity: pipeline, as: new, effect: reads, facts: []}
  contexts:
    web: {place: 'web::signed-in::manual-pipeline'}
- text: The Product creates the new Pipeline with its Workflows and Steps
  kind: product
  actor: user
  capability: run-pipeline
  entities:
  - {entity: pipeline, as: new, effect: creates, to: Pending, facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Additional pipeline variables]}
  - {entity: workflow, effect: creates, to: Pending, facts: [Name, Platform, Matrix variables, Depends on], with: new}
  - {entity: step, effect: creates, to: Pending, facts: [Name, Type], with: workflow}
  contexts:
    web: {place: 'web::signed-in::manual-pipeline'}
- text: The Product opens the new Pipeline
  kind: product
  actor: user
  capability: run-pipeline
  entities:
  - {entity: pipeline, as: new, effect: reads, facts: [Number]}
  contexts:
    web: {place: 'web::signed-in::pipeline'}
- text: The User chooses a Step of the new Pipeline and follows its log
  kind: actor
  actor: user
  capability: view-pipeline
  entities:
  - {entity: pipeline, as: new, effect: reads, facts: [Number]}
  - {entity: step, effect: reads, facts: [Name, Log]}
  contexts:
    web: {place: 'web::signed-in::pipeline'}
---

# Run a pipeline and follow it

## Trigger

A person runs a pipeline by hand.

## Outcome

The person follows the pipeline they started.
