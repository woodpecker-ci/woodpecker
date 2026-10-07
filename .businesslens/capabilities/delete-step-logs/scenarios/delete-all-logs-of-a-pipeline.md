---
kind: alternative
routes: {cli: CLI, api: API}
steps:
- text: The User purges the logs of a finished Pipeline
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number]
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the log of every Step of the Pipeline
  kind: product
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number]
  - entity: step
    effect: changes
    facts: [Log]
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Delete every log of a pipeline

## Trigger

Every log of a pipeline must go.

## Outcome

No step of the pipeline has a log.

## Edge cases

- A pending, running or blocked pipeline's logs cannot be deleted.
