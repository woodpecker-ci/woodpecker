---
kind: alternative
result: achieved
routes: {web: Web}
steps:
- text: The User chooses Run now on a Cron
  kind: actor
  actor: user
  capability: run-cron
  entities:
  - entity: cron
    effect: reads
    facts: [Name, Branch, Variables]
  contexts:
    web: {place: 'web::signed-in::crons'}
- text: The Product creates the cron Pipeline with its Workflows and Steps
  kind: product
  actor: user
  capability: run-cron
  entities:
  - entity: pipeline
    as: new
    effect: creates
    to: Pending
    facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Cron, Additional pipeline variables]
  - entity: workflow
    effect: creates
    to: Pending
    facts: [Name, Platform, Matrix variables, Depends on]
    with: new
  - entity: step
    effect: creates
    to: Pending
    facts: [Name, Type]
    with: workflow
  - entity: cron
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::crons'}
- text: The Product opens the new Pipeline
  kind: product
  actor: user
  capability: run-cron
  entities:
  - entity: pipeline
    as: new
    effect: reads
    facts: [Number]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
- text: The User chooses a Step of the new Pipeline and follows its log
  kind: actor
  actor: user
  capability: view-pipeline
  entities:
  - entity: pipeline
    as: new
    effect: reads
    facts: [Number]
  - entity: step
    effect: reads
    facts: [Name, Log]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
---

# Run a cron now and follow it

## Trigger

A person runs a cron now.

## Outcome

The person follows the cron pipeline they started.
