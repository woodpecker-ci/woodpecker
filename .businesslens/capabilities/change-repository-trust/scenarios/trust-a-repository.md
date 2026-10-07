---
kind: primary
routes:
  web: Web
  cli: CLI
  api: API
steps:
- text: The Administrator turns the trusted network, volumes or security settings on or off
  kind: actor
  actor: administrator
  entities: []
  contexts:
    web: {place: 'web::signed-in::project-settings'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Repository's trusted settings
  kind: product
  actor: administrator
  entities:
  - {entity: repository, effect: changes, facts: [Trusted]}
  contexts:
    web: {place: 'web::signed-in::project-settings'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Trust a repository

## Trigger

A repository's pipelines need privileged options.

## Outcome

Pipelines of the repository may use the trusted options that are on.

## Edge cases

- A repository admin who is not an administrator is refused when the trusted settings would change; sending them unchanged is accepted.
