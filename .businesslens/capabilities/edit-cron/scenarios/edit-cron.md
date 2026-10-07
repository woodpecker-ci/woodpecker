---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User changes a Cron and chooses Save cron
  kind: actor
  actor: user
  entities:
  - entity: cron
    effect: reads
    facts: [Name]
  contexts:
    web: {place: 'web::signed-in::crons'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Cron and works out its next execution
  kind: product
  actor: user
  entities:
  - entity: cron
    effect: changes
    facts: [Name, Schedule, Timezone, Branch, Enabled, Variables, Next execution]
  contexts:
    web: {place: 'web::signed-in::crons'}
    api: {place: 'api::signed-in'}
---

# Edit a cron

## Trigger

A schedule needs to change or pause.

## Outcome

The cron runs on its new schedule, or not at all while disabled.
