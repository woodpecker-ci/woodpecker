---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User opens a Repository
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::signed-in::repository'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product lists its pipelines newest first
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name, Forge link]
  - entity: pipeline
    effect: reads
    facts: [Number, Event, Branch, Message, Author, Created, Duration]
  contexts:
    web: {place: 'web::signed-in::repository'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Browse a repository's pipelines

## Trigger

A signed-in person opens a repository or lists its pipelines in the CLI.

## Outcome

They see the repository's activity.

## Edge cases

- Choosing a branch or pull request lists only its pipelines.
- A repository the person may not read answers as not found.
