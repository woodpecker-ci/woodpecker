---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User chooses Run now on a Cron
  kind: actor
  actor: user
  entities:
  - entity: cron
    effect: reads
    facts: [Name, Branch, Variables]
  contexts:
    web: {place: 'web::signed-in::crons'}
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
    web: {place: 'web::signed-in::crons'}
    api: {place: 'api::signed-in'}
- text: The Product creates a Pipeline in Pending with a Workflow for each configuration file and the Steps it defines
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: creates
    to: Pending
    facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Cron, Additional pipeline variables]
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
    web: {place: 'web::signed-in::crons'}
    api: {place: 'api::signed-in'}
- text: The Product opens the new Pipeline
  kind: product
  actor: user
  entities:
  - {entity: pipeline, effect: reads, facts: [Number]}
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    api: {place: 'api::signed-in'}
---

# Run a cron now

## Trigger

A person wants a scheduled pipeline right away.

## Outcome

A cron pipeline is queued; the schedule is unchanged.
