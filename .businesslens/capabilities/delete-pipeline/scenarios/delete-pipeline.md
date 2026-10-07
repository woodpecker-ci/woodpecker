---
kind: primary
routes: {cli: CLI, api: API}
steps:
- text: The User deletes a finished Pipeline
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number]
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Pipeline with its Workflows and Steps
  kind: product
  actor: user
  entities:
  - {entity: pipeline, effect: removes, from: Success}
  - {entity: workflow, effect: removes, from: Success, with: pipeline}
  - {entity: step, effect: removes, from: Success, with: workflow}
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Delete a pipeline

## Trigger

Old or unwanted pipelines clutter a repository.

## Outcome

The pipeline is gone from the repository's activity.

## Edge cases

- A pending, running or blocked pipeline cannot be deleted.
- Purging with a dry run lists what would be deleted and deletes nothing.
