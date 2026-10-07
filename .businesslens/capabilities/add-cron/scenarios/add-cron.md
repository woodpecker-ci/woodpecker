---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User enters a name, schedule, timezone, branch and variables and chooses Save cron
  kind: actor
  actor: user
  entities:
  - entity: cron
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::crons'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product creates the Cron for the Repository with its next execution
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  - entity: cron
    effect: creates
    facts: [Name, Schedule, Timezone, Branch, Enabled, Variables, Next execution]
  contexts:
    web: {place: 'web::signed-in::crons'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Add a cron

## Trigger

A repository needs pipelines on a schedule.

## Outcome

The cron shows when it runs next.

## Edge cases

- A schedule that does not parse or never runs again is refused, and so is a branch the forge cannot resolve.
- Names are unique within the repository.
- The CLI sets no timezone or variables: such a cron runs in UTC with none.
