---
kind: validation
routes: {cli: CLI, api: API}
steps:
- text: The User deploys a Pipeline
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    effect: reads
    facts: [Number]
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Repository does not allow deployments
  kind: condition
  entities:
  - entity: repository
    effect: reads
    facts: [Allow deployments]
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
  actor: user
- text: 'The Product refuses: the repository does not allow deployments'
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: []
  contexts:
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Refuse a deployment the repository does not allow

## Trigger

A person deploys in a repository with deployments off.

## Outcome

No deployment pipeline is created.
