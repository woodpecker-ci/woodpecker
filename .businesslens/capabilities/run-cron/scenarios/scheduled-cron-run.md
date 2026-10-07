---
kind: unattended
routes:
  web: Web
steps:
- text: An enabled Cron of an enabled Repository reaches its next execution
  kind: condition
  unattended: true
  entities:
  - {entity: cron, effect: reads, facts: [Next execution, Enabled, Schedule, Timezone]}
  - {entity: repository, effect: reads, facts: [Default branch]}
  contexts:
    web: {place: 'web::signed-in::repository'}
- text: The Product moves the Cron's next execution to its next scheduled time
  kind: product
  entities:
  - {entity: cron, effect: changes, facts: [Next execution]}
  contexts:
    web: {place: 'web::signed-in::repository'}
- text: The Product reads the pipeline configuration of the Repository at that commit
  kind: product
  entities:
  - {entity: repository, effect: reads, facts: [Pipeline path, Config extension endpoint, Config extension exclusive]}
  - {entity: pipeline, effect: reads, facts: []}
  contexts:
    web: {place: 'web::signed-in::repository'}
- text: The Product creates a Pipeline in Pending with a Workflow for each configuration file and the Steps it defines
  kind: product
  entities:
  - {entity: pipeline, effect: creates, to: Pending, facts: [Number, Event, Branch, Commit, Message, Author, Created, Configuration, Woodpecker version, Cron, Additional pipeline variables]}
  - {entity: workflow, effect: creates, to: Pending, facts: [Name, Platform, Matrix variables, Depends on], with: pipeline}
  - {entity: step, effect: creates, to: Pending, facts: [Name, Type], with: workflow}
  contexts:
    web: {place: 'web::signed-in::repository'}
---

# Run a cron on schedule

## Trigger

A cron's scheduled time passes.

## Outcome

A cron pipeline shows in the repository's activity.

## Edge cases

- Each scheduled time starts one pipeline, even when several servers share a database.
