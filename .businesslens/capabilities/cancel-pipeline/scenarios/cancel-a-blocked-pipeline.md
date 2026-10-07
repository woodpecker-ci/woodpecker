---
kind: alternative
routes: {cli: CLI, api: API}
steps:
- text: The User cancels a Pipeline that awaits approval
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number]
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product marks the Pipeline killed and canceled by the User, leaving its workflows and steps awaiting approval
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: changes
    from: Blocked
    to: Killed
    facts: [Cancel info]
  - entity: step
    effect: reads
    facts: []
  - entity: workflow
    effect: reads
    facts: []
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Cancel a pipeline awaiting approval

## Trigger

A pipeline awaiting approval is no longer wanted, and nobody will approve or decline it.

## Outcome

The pipeline is killed without running; the web UI offers no Cancel while a pipeline awaits approval.
