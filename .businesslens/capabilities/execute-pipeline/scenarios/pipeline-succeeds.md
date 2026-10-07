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
  - {entity: step, effect: reads, facts: []}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: The Product starts each Step in order and streams its log
  kind: product
  entities:
  - {entity: step, effect: changes, from: Pending, to: Running, facts: [Log]}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: Every Step exits with code 0
  kind: product
  entities:
  - {entity: step, effect: changes, from: Running, to: Success, facts: [Exit code, Duration]}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
- text: The Product finishes the Workflow and the Pipeline successfully and reports the result to the Forge
  kind: product
  entities:
  - {entity: workflow, effect: changes, from: Running, to: Success, facts: []}
  - {entity: pipeline, effect: changes, from: Running, to: Success, facts: [Duration]}
  - {entity: forge, effect: reads, facts: [URL]}
  contexts:
    public: {place: 'web::public::pipeline'}
    signed-in: {place: 'web::signed-in::pipeline'}
---

# Run a pipeline to success

## Trigger

A workflow is queued.

## Outcome

The pipeline succeeded and the forge shows it.

## Edge cases

- Workflows wait for those they depend on before they are given to an agent.
- While the queue is paused no workflow is given to an agent.
