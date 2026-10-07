---
kind: unattended
routes:
  public: Public pipeline
  signed-in: Pipeline
steps:
- text: An Agent with free capacity asks for work while the Queue is running
  kind: condition
  unattended: true
  entities:
  - {entity: agent, effect: reads, facts: [Capacity, Custom labels, Filters, Disabled]}
  - {entity: queue, effect: reads, facts: [Tasks]}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: The Product gives the Agent the oldest pending Workflow whose labels and organization it matches, and starts the Pipeline
  kind: product
  entities:
  - {entity: workflow, effect: changes, from: Pending, to: Running, facts: [Agent, Platform]}
  - {entity: pipeline, effect: changes, from: Pending, to: Running, facts: [Duration]}
  - {entity: agent, effect: changes, facts: [Last contact]}
  - {entity: organization, effect: reads, facts: []}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: The Product hands each Step the Secrets and Registry credentials its events and plugins allow
  kind: product
  entities:
  - {entity: secret, effect: reads, facts: [Value, Available for plugins, Available at events]}
  - {entity: registry, effect: reads, facts: [Address, Username, Password]}
  - {entity: step, as: failed, effect: reads, facts: []}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: The Product starts each Step in order and streams its log
  kind: product
  entities:
  - {entity: step, as: failed, effect: changes, from: Pending, to: Running, facts: [Log]}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: A Step exits with another code
  kind: product
  entities:
  - {entity: step, as: failed, effect: changes, from: Running, to: Failure, facts: [Exit code, Duration]}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: The Product skips the Steps that only run on success
  kind: product
  entities:
  - {entity: step, as: later, effect: changes, from: Pending, to: Skipped, facts: []}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: The Product fails the Workflow and the Pipeline and reports the result to the Forge
  kind: product
  entities:
  - {entity: workflow, effect: changes, from: Running, to: Failure, facts: [Error]}
  - {entity: pipeline, effect: changes, from: Running, to: Failure, facts: [Duration]}
  - {entity: forge, effect: reads, facts: [URL]}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
---

# Run a pipeline that fails

## Trigger

A step of a queued workflow fails.

## Outcome

The pipeline failed and the forge shows it.

## Edge cases

- A step allowed to fail does not fail its workflow.
- A workflow running longer than the repository's timeout is stopped.
- Workflows that depend on a failed workflow are skipped.
- A step asking for a secret not available to its event or image ends the pipeline in error, naming the secret.
- A step whose agent cannot run it fails with the reason recorded on the step and its workflow.
