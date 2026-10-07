---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The Visitor opens a public Repository
  kind: actor
  actor: visitor
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::public::repository'}
    api: {place: 'api::public'}
- text: The Product lists its pipelines newest first
  kind: product
  actor: visitor
  entities:
  - entity: repository
    effect: reads
    facts: [Full name, Forge link]
  - entity: pipeline
    effect: reads
    facts: [Number, Event, Branch, Message, Author, Created, Duration]
  contexts:
    web: {place: 'web::public::repository'}
    api: {place: 'api::public'}
---

# Browse a public repository's pipelines

## Trigger

Someone not signed in opens a public repository.

## Outcome

They see the repository's activity.

## Edge cases

- A private or internal repository is not shown; the visitor is asked to sign in.
