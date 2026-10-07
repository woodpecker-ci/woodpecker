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
- text: The Product reads the pipeline configuration of the Repository at that commit
  kind: product
  actor: forge
  entities:
  - {entity: repository, effect: reads, facts: [Pipeline path, Config extension endpoint, Config extension exclusive]}
  - {entity: pipeline, as: newer, effect: reads, facts: []}
  contexts:
    hook: {place: forge-webhook}
- text: The Product creates a Pipeline in Pending with a Workflow for each configuration file and the Steps it defines
  kind: product
  actor: forge
  entities:
  - {entity: pipeline, as: newer, effect: creates, to: Pending, facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Changed files]}
  - {entity: workflow, effect: creates, to: Pending, facts: [Name, Platform, Matrix variables, Depends on], with: newer}
  - {entity: step, effect: creates, to: Pending, facts: [Name, Type], with: workflow}
  contexts:
    hook: {place: forge-webhook}
- text: The Repository cancels previous pipelines for this event and an earlier Pipeline of the same branch or ref is still running
  kind: condition
  entities:
  - {entity: repository, effect: reads, facts: [Cancel previous pipelines]}
  - {entity: pipeline, as: earlier, effect: reads, facts: [Branch]}
  contexts:
    hook: {place: forge-webhook}
  actor: forge
- text: The Product kills the earlier Pipeline as superseded by the new one
  kind: product
  actor: forge
  entities:
  - {entity: pipeline, as: earlier, effect: changes, from: Running, to: Killed, facts: [Cancel info]}
  contexts:
    hook: {place: forge-webhook}
---

# Supersede an earlier pipeline

## Trigger

A new push arrives while the previous pipeline of the branch still runs.

## Outcome

Only the newest pipeline keeps running; the earlier one says which pipeline superseded it.

## Edge cases

- A pending or blocked earlier pipeline is canceled the same way; push pipelines match by branch, others by ref.
