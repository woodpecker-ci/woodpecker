---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User chooses Delete repository and confirms that all data will be lost
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::signed-in::repository-actions'}
    api: {place: 'api::signed-in'}
- text: The Product removes the webhook and deletes the Repository with its pipelines, secrets, registries and permissions
  kind: product
  actor: user
  entities:
  - {entity: repository, effect: removes, from: Enabled}
  - {entity: pipeline, effect: removes, from: Success, with: repository}
  - {entity: secret, effect: removes, with: repository}
  - {entity: registry, effect: removes, with: repository}
  - {entity: repository-permission, effect: removes, with: repository}
  contexts:
    web: {place: 'web::signed-in::repository-actions'}
    api: {place: 'api::signed-in'}
- text: The Product returns the User to the repository list
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repositories'}
    api: {place: 'api::signed-in'}
---

# Delete a repository

## Trigger

A repository admin wants a repository gone from Woodpecker.

## Outcome

Nothing of the repository remains in Woodpecker.
