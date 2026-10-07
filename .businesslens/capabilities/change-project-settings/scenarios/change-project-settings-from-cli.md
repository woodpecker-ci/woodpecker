---
kind: alternative
routes: {cli: CLI}
steps:
- text: The User changes the project settings and chooses Save settings
  kind: actor
  actor: user
  entities: []
  contexts:
    cli: {place: 'cli::signed-in'}
- text: The Product saves the Repository's project settings
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: changes
    facts: [Pipeline path, Project visibility, Approval requirements, Timeout]
  contexts:
    cli: {place: 'cli::signed-in'}
---

# Change project settings from the CLI

## Trigger

A repository admin wants pipelines to behave differently.

## Outcome

Pipelines started afterwards follow the new settings.

## Edge cases

- An unknown visibility or approval requirement is refused.
- A timeout below one minute is refused.
