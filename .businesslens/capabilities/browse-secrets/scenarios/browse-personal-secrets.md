---
kind: alternative
routes: {web: Web}
steps:
- text: The User opens their secrets
  kind: actor
  actor: user
  entities:
  - entity: secret
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::user-secrets'}
- text: The Product lists the Secrets of the User's personal organization
  kind: product
  actor: user
  entities:
  - entity: organization
    effect: reads
    facts: [Name]
  - entity: secret
    effect: reads
    facts: [Name, Note, Available for plugins, Available at events]
  contexts:
    web: {place: 'web::signed-in::user-secrets'}
---

# Browse personal secrets

## Trigger

A person checks their own secrets.

## Outcome

The person sees their secrets.
