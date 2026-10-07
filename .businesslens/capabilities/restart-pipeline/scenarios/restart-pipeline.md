---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Restart on a Pipeline
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    as: earlier
    effect: reads
    facts: [Number]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product creates a new Pipeline in Pending from the earlier one's configuration, with a Workflow and Steps for each file
  kind: product
  actor: user
  entities:
  - entity: pipeline
    as: restart
    effect: creates
    to: Pending
    facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Restarted from, Additional pipeline variables]
  - entity: workflow
    effect: creates
    to: Pending
    facts: [Name, Platform, Matrix variables, Depends on]
    with: restart
  - entity: step
    effect: creates
    to: Pending
    facts: [Name, Type]
    with: workflow
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product opens the new Pipeline
  kind: product
  actor: user
  entities:
  - entity: pipeline
    as: restart
    effect: reads
    facts: [Number]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Restart a pipeline

## Trigger

A pipeline failed for a reason outside its code, or needs to run again.

## Outcome

A new pipeline runs and the person follows it.

## Edge cases

- A blocked pipeline cannot be restarted; it is approved or declined.
- When the configuration can no longer be found, the new pipeline ends in Error.
- Restarting a declined pipeline runs it without asking for approval again.
- When a configuration extension fails, the restart is refused and no pipeline is created.
