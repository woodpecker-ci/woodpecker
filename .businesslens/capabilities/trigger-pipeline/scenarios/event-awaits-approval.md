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
- text: The Repository requires approval for this event and its author is not an allowed user
  kind: condition
  entities:
  - {entity: repository, effect: reads, facts: [Approval requirements, Allowed users]}
  - {entity: user, effect: reads, facts: []}
  contexts:
    hook: {place: forge-webhook}
  actor: forge
- text: The Product reads the pipeline configuration of the Repository at that commit
  kind: product
  actor: forge
  entities:
  - {entity: repository, effect: reads, facts: [Pipeline path, Config extension endpoint, Config extension exclusive]}
  - {entity: pipeline, effect: reads, facts: []}
  contexts:
    hook: {place: forge-webhook}
- text: The Product creates a Pipeline in Blocked with a Workflow for each configuration file and the Steps it defines
  kind: product
  actor: forge
  entities:
  - {entity: pipeline, effect: creates, to: Blocked, facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Changed files]}
  - {entity: workflow, effect: creates, to: Blocked, facts: [Name, Platform, Matrix variables, Depends on], with: pipeline}
  - {entity: step, effect: creates, to: Blocked, facts: [Name, Type], with: workflow}
  contexts:
    hook: {place: forge-webhook}
- text: The Product reports the Pipeline as awaiting approval to the Forge
  kind: product
  actor: forge
  entities:
  - {entity: pipeline, effect: reads, facts: [Number]}
  contexts:
    hook: {place: forge-webhook}
---

# Hold a pipeline for approval

## Trigger

A pull request from a fork, or another event the repository requires approval for, arrives.

## Outcome

The pipeline waits until someone with push access approves or declines it.
