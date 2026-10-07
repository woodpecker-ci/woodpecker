---
kind: validation
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User sets a timeout above the server maximum
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::project-settings'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product refuses the change and names the maximum timeout
  kind: product
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::project-settings'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Refuse a timeout above the maximum

## Trigger

A repository admin who is not an administrator raises the timeout above the maximum.

## Outcome

The settings stay as they were.
