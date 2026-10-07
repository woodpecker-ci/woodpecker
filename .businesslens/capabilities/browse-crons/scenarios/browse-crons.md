---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User opens the crons of a Repository
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: cron
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::crons'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product lists the Repository's Crons with their next execution
  kind: product
  actor: user
  entities:
  - entity: cron
    effect: reads
    facts: [Name, Schedule, Timezone, Branch, Enabled, Variables, Next execution]
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::crons'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Browse a repository's crons

## Trigger

A person checks what runs on a schedule.

## Outcome

The person sees each cron, whether it is enabled and when it runs next.

## Edge cases

- A disabled cron shows as disabled instead of its next execution.
