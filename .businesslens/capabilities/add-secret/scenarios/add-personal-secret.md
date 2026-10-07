---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User enters a name, value, note, events and plugins and chooses Add
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::user-secrets'}
    api: {place: 'api::signed-in'}
- text: The Product creates the Secret for the User's personal organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: secret
    effect: creates
    facts: [Name, Value, Note, Available for plugins, Available at events]
  contexts:
    web: {place: 'web::signed-in::user-secrets'}
    api: {place: 'api::signed-in'}
---

# Add a personal secret

## Trigger

Pipelines of their personal organization need a secret.

## Outcome

Pipelines of their personal organization receive the secret.

## Edge cases

- Names are unique per repository, organization or server; saving a duplicate fails.
