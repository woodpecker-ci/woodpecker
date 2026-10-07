---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User opens the Pipeline's debug information and chooses Download metadata
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Metadata]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    api: {place: 'api::signed-in'}
---

# Download pipeline metadata

## Trigger

A person wants to reproduce a pipeline locally.

## Outcome

They have a metadata file the CLI can run the pipeline with.

## Edge cases

- Without push access the debug information is not shown.
