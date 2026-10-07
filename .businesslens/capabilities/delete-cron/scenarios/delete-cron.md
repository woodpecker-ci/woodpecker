---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User chooses Delete cron
  kind: actor
  actor: user
  entities:
  - entity: cron
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::crons'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product deletes the Cron
  kind: product
  actor: user
  entities:
  - {entity: cron, effect: removes}
  contexts:
    web: {place: 'web::signed-in::crons'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Delete a cron

## Trigger

A schedule is no longer needed.

## Outcome

No more pipelines start from it.
