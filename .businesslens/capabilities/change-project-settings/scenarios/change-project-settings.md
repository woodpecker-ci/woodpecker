---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User changes the project settings and chooses Save settings
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::project-settings'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Repository's project settings
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: changes
    facts: [Pipeline path, Project visibility, Allow pull requests, Allow deployments, Approval requirements, Allowed users, Timeout, Cancel previous pipelines, Custom trusted clone plugins]
  contexts:
    web: {place: 'web::signed-in::project-settings'}
    api: {place: 'api::signed-in'}
---

# Change project settings

## Trigger

A repository admin wants pipelines to behave differently.

## Outcome

Pipelines started afterwards follow the new settings.

## Edge cases

- An unknown visibility or approval requirement is refused.
- A timeout below one minute is refused.
