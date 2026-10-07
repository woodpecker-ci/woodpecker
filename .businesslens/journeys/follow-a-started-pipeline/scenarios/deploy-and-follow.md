---
kind: alternative
result: achieved
routes:
  web: Web
steps:
- text: The User chooses Deploy on a successful Pipeline and enters the target environment
  kind: actor
  actor: user
  capability: deploy-pipeline
  entities:
  - {entity: pipeline, as: earlier, effect: reads, facts: [Number]}
  contexts:
    web: {place: 'web::signed-in::pipeline'}
- text: The Product creates the deployment Pipeline with its Workflows and Steps
  kind: product
  actor: user
  capability: deploy-pipeline
  entities:
  - {entity: pipeline, as: new, effect: creates, to: Pending, facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Restarted from, Deployment target, Deployment task, Additional pipeline variables]}
  - {entity: workflow, effect: creates, to: Pending, facts: [Name, Platform, Matrix variables, Depends on], with: new}
  - {entity: step, effect: creates, to: Pending, facts: [Name, Type], with: workflow}
  contexts:
    web: {place: 'web::signed-in::pipeline'}
- text: The Product opens the deployment Pipeline
  kind: product
  actor: user
  capability: deploy-pipeline
  entities:
  - {entity: pipeline, as: new, effect: reads, facts: [Number]}
  contexts:
    web: {place: 'web::signed-in::pipeline'}
- text: The User chooses a Step of the new Pipeline and follows its log
  kind: actor
  actor: user
  capability: view-pipeline
  entities:
  - {entity: pipeline, as: new, effect: reads, facts: [Number]}
  - {entity: step, effect: reads, facts: [Name, Log]}
  contexts:
    web: {place: 'web::signed-in::pipeline'}
---

# Deploy a pipeline and follow it

## Trigger

A person deploys a successful pipeline.

## Outcome

The person follows the deployment.
