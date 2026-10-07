---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Visitor opens a Pipeline of a public repository
  kind: actor
  actor: visitor
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number]
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::public::pipeline'}
    api: {place: 'api::public'}
- text: The Product shows its workflows and steps
  kind: product
  actor: visitor
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
    web: {place: 'web::public::pipeline'}
    api: {place: 'api::public'}
- text: The Visitor chooses a Step and the Product streams its log
  kind: actor
  actor: visitor
  entities:
  - entity: step
    effect: reads
    facts: [Name, Exit code, Duration, Log]
  contexts:
    web: {place: 'web::public::pipeline'}
    api: {place: 'api::public'}
---

# View a public pipeline

## Trigger

Someone not signed in follows a public pipeline.

## Outcome

They see the pipeline's progress and logs live.
