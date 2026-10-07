---
kind: validation
routes:
  hook: Forge webhook
steps:
- text: The Forge sends a pull request event for an enabled Repository
  kind: actor
  actor: forge
  entities:
  - {entity: repository, effect: reads, facts: [Full name]}
  contexts:
    hook: {place: forge-webhook}
- text: The Repository does not allow pull requests
  kind: condition
  entities:
  - {entity: repository, effect: reads, facts: [Allow pull requests]}
  contexts:
    hook: {place: forge-webhook}
  actor: forge
- text: The Product ignores the event
  kind: product
  actor: forge
  entities: []
  contexts:
    hook: {place: forge-webhook}
---

# Ignore pull requests the repository does not allow

## Trigger

A pull request is opened in a repository with pull requests off.

## Outcome

No pipeline is created.
