---
kind: edge
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
- text: The configuration cannot be loaded or does not parse
  kind: condition
  entities: []
  contexts:
    hook: {place: forge-webhook}
  actor: forge
- text: The Product creates the Pipeline in Error with the parse errors and reports it to the Forge
  kind: product
  actor: forge
  entities:
  - {entity: pipeline, effect: creates, to: Error, facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Errors]}
  contexts:
    hook: {place: forge-webhook}
---

# End in error on invalid configuration

## Trigger

A commit breaks the pipeline configuration.

## Outcome

The pipeline shows its errors and the forge shows it as failed.
