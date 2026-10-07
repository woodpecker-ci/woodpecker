---
kind: alternative
routes: {web: Web}
steps:
- text: The User chooses a branch, events, and optionally a workflow and a step, and the syntax of the link
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name, Default branch]
  - entity: step
    effect: reads
    facts: []
  - entity: workflow
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::badge'}
- text: The Product shows the badge of the latest matching Pipeline and the link to embed it
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Event, Branch]
  contexts:
    web: {place: 'web::signed-in::badge'}
---

# Build a badge link

## Trigger

A repository admin wants the repository's status in a README.

## Outcome

The admin has a URL, Markdown or HTML snippet showing the badge.
