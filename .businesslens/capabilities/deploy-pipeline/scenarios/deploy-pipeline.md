---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Deploy on a successful Pipeline and enters the target environment, task and variables
  kind: actor
  actor: user
  entities:
  - entity: pipeline
    as: earlier
    effect: reads
    facts: [Number]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Repository allows deployments
  kind: condition
  entities:
  - entity: repository
    effect: reads
    facts: [Allow deployments]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
  actor: user
- text: The Product creates a deployment Pipeline in Pending from the earlier one's configuration
  kind: product
  actor: user
  entities:
  - entity: pipeline
    as: deployment
    effect: creates
    to: Pending
    facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Restarted from, Deployment target, Deployment task, Additional pipeline variables]
  - entity: workflow
    effect: creates
    to: Pending
    facts: [Name, Platform, Matrix variables, Depends on]
    with: deployment
  - entity: step
    effect: creates
    to: Pending
    facts: [Name, Type]
    with: workflow
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product opens the deployment Pipeline
  kind: product
  actor: user
  entities:
  - entity: pipeline
    as: deployment
    effect: reads
    facts: [Number]
  contexts:
    web: {place: 'web::signed-in::pipeline'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Deploy a pipeline

## Trigger

A person wants a built commit deployed.

## Outcome

A deployment pipeline runs and the person follows it.
