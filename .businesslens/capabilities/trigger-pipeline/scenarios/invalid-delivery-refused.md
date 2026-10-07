---
kind: validation
routes:
  hook: Forge webhook
steps:
- text: The Forge sends an event whose token does not belong to the Repository it names
  kind: actor
  actor: forge
  entities:
  - {entity: repository, effect: reads, facts: [Webhook]}
  contexts:
    hook: {place: forge-webhook}
- text: The Product refuses the event and starts nothing
  kind: product
  actor: forge
  entities: []
  contexts:
    hook: {place: forge-webhook}
---

# Refuse an event that does not belong to its repository

## Trigger

A webhook call carries a missing or foreign token, or names a different repository than its token.

## Outcome

No pipeline is created.
