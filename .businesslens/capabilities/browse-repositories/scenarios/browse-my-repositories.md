---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User opens the repository list
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repositories'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product lists each enabled Repository the User may read, with its latest Pipeline
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name, Project visibility]
  - entity: pipeline
    effect: reads
    facts: [Number, Event, Branch, Message]
  contexts:
    web: {place: 'web::signed-in::repositories'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Browse my repositories

## Trigger

A signed-in person opens Woodpecker or lists repositories in the CLI.

## Outcome

The person sees the repositories they can work with.

## Edge cases

- A search narrows the list by name.
