---
kind: alternative
routes: {web: Web, api: API}
steps:
- text: The User chooses Enable repository
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::signed-in::repository-actions'}
    api: {place: 'api::signed-in'}
- text: The Product installs the webhook again and enables the Repository, owned by the User, keeping its settings
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: changes
    from: Disabled
    to: Enabled
    facts: [Owner, Webhook]
  contexts:
    web: {place: 'web::signed-in::repository-actions'}
    api: {place: 'api::signed-in'}
---

# Enable a disabled repository again

## Trigger

A person wants a disabled repository built again.

## Outcome

The repository is enabled with its earlier pipelines, secrets and settings.
