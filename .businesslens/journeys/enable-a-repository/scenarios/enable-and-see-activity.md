---
kind: primary
result: achieved
routes:
  web: Web
steps:
- text: The User picks a forge repository to enable
  kind: actor
  actor: user
  capability: enable-repository
  entities:
  - {entity: forge, effect: reads, facts: []}
  - {entity: repository, effect: reads, facts: []}
  contexts:
    web: {place: 'web::signed-in::add-repository'}
- text: The Product creates the Repository as Enabled and installs its webhook
  kind: product
  actor: user
  capability: enable-repository
  entities:
  - {entity: repository, effect: creates, to: Enabled, facts: [Full name, Forge link, Default branch, Project visibility, Approval requirements, Allow pull requests, Allow deployments, Cancel previous pipelines, Timeout, Owner, Webhook]}
  - {entity: repository-permission, effect: creates, facts: [Pull, Push, Admin, Synced], with: repository}
  contexts:
    web: {place: 'web::signed-in::add-repository'}
- text: The Product opens the new Repository
  kind: product
  actor: user
  capability: enable-repository
  entities:
  - {entity: repository, effect: reads, facts: [Full name]}
  contexts:
    web: {place: 'web::signed-in::repository'}
- text: The User looks through the Repository's pipelines
  kind: actor
  actor: user
  capability: browse-pipelines
  entities:
  - {entity: repository, effect: reads, facts: [Full name]}
  - {entity: pipeline, effect: reads, facts: [Number, Event, Branch]}
  contexts:
    web: {place: 'web::signed-in::repository'}
---

# Enable a repository and see its activity

## Trigger

A person enables a repository from Add repository.

## Outcome

The person is on the repository's activity, ready for its first pipeline.
