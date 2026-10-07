---
kind: primary
routes:
  api: API
steps:
- text: The Visitor requests the status badge of a Repository
  kind: actor
  actor: visitor
  entities:
  - {entity: repository, effect: reads, facts: [Default branch]}
  contexts:
    api: {place: 'api::public'}
- text: The Product answers with the state of its latest matching Pipeline
  kind: product
  actor: visitor
  entities:
  - {entity: pipeline, effect: reads, facts: [Event, Branch]}
  contexts:
    api: {place: 'api::public'}
---

# View a status badge

## Trigger

A README or dashboard shows a repository's build state.

## Outcome

The badge shows the latest pipeline's state, or that there is none.

## Edge cases

- An unknown event in the request is refused.
- With a workflow, and optionally a step, the badge shows that workflow's or step's state.
- The CCMenu feed of the same repository names the latest pipeline, its number, state and start time.
