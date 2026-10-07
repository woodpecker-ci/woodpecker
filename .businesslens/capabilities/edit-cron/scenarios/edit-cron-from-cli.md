---
kind: alternative
routes: {cli: CLI}
steps:
- text: The User changes a Cron and chooses Save cron
  kind: actor
  actor: user
  entities:
  - entity: cron
    effect: reads
    facts: [Name]
  contexts:
    cli: {place: 'cli::signed-in'}
- text: The Product saves the Cron and works out its next execution
  kind: product
  actor: user
  entities:
  - entity: cron
    effect: changes
    facts: [Name, Schedule, Branch, Enabled, Next execution]
  contexts:
    cli: {place: 'cli::signed-in'}
---

# Edit a cron from the CLI

## Trigger

A schedule needs to change or pause.

## Outcome

The cron runs on its new schedule, or not at all while disabled.
