---
kind: unattended
routes: {public: Public pipeline, signed-in: Pipeline}
steps:
- text: A Workflow ends in failure while another Workflow of the Pipeline depends on it
  kind: condition
  unattended: true
  entities:
  - entity: workflow
    as: failed
    effect: reads
    facts: [Name]
  - entity: workflow
    as: dependent
    effect: reads
    facts: [Depends on]
  - entity: pipeline
    effect: reads
    facts: []
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: The Product skips the dependent Workflow without giving it to an Agent and kills its Steps
  kind: product
  entities:
  - entity: workflow
    as: dependent
    effect: changes
    from: Pending
    to: Skipped
    facts: []
  - entity: step
    effect: changes
    from: Pending
    to: Killed
    facts: []
  - entity: agent
    effect: reads
    facts: []
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
---

# Skip a workflow whose dependency failed

## Trigger

A workflow fails and another workflow runs only after it succeeds.

## Outcome

The dependent workflow shows as skipped and the pipeline fails.

## Edge cases

- A workflow whose status conditions accept failure runs after the failed workflow instead.
- A required dependency that matches no workflow drops the dependent workflow from the pipeline; an optional one is ignored.
