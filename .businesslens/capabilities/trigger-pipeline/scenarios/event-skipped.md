---
kind: alternative
routes:
  hook: Forge webhook
steps:
- text: The Forge sends an event for an enabled Repository
  kind: actor
  actor: forge
  entities:
  - {entity: repository, effect: reads, facts: [Full name]}
  contexts:
    hook: {place: forge-webhook}
- text: The commit message asks to skip CI, or no workflow's conditions match the event
  kind: condition
  entities:
  - {entity: workflow, effect: reads, facts: []}
  contexts:
    hook: {place: forge-webhook}
  actor: forge
- text: The Product starts nothing
  kind: product
  actor: forge
  entities: []
  contexts:
    hook: {place: forge-webhook}
---

# Skip an event that asks for no pipeline

## Trigger

A commit carries [skip ci], or no workflow runs for the event.

## Outcome

Nothing is queued.

## Edge cases

- No configuration file exists at the pipeline path: nothing is queued either.
- [skip ci] in the commit message counts only for push and pull request events.
