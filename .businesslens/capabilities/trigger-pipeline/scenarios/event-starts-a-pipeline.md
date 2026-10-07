---
kind: primary
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
  - {entity: pipeline, effect: reads, facts: []}
  contexts:
    hook: {place: forge-webhook}
- text: The Product creates a Pipeline in Pending with a Workflow for each configuration file and the Steps it defines
  kind: product
  actor: forge
  entities:
  - {entity: pipeline, effect: creates, to: Pending, facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Changed files]}
  - {entity: workflow, effect: creates, to: Pending, facts: [Name, Platform, Matrix variables, Depends on], with: pipeline}
  - {entity: step, effect: creates, to: Pending, facts: [Name, Type], with: workflow}
  contexts:
    hook: {place: forge-webhook}
- text: The Product reports the Pipeline as pending to the Forge
  kind: product
  actor: forge
  entities:
  - {entity: pipeline, effect: reads, facts: [Number]}
  contexts:
    hook: {place: forge-webhook}
---

# Start a pipeline from a forge event

## Trigger

Someone pushes, tags, releases, opens or updates a pull request, or deploys on the forge.

## Outcome

A pipeline is queued and the forge shows it as pending.

## Edge cases

- Events of a disabled repository, or of a repository without an owner, are ignored.
- A pipeline that takes longer than the webhook timeout to create is answered as accepted and created in the background.
