---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User opens a Pipeline
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
- text: The Product shows its workflows and steps
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number, Event, Branch, Commit, Message, Author, Duration, Errors, Configuration, Changed files]
  - entity: workflow
    effect: reads
    facts: [Name, Platform, Agent, Error]
  - entity: step
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The User chooses a Step and the Product streams its log
  kind: actor
  actor: user
  entities:
  - entity: step
    effect: reads
    facts: [Name, Exit code, Duration, Log]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# View a pipeline

## Trigger

A signed-in person follows a pipeline.

## Outcome

They see the pipeline's progress, logs, configuration, changed files and errors.

## Edge cases

- Linter warnings are shown together with the pipeline's errors.
- Logs longer than the server's maximum line count are cut off in the view and can be downloaded whole.
