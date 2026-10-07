---
kind: primary
routes:
  web: Web
steps:
- text: The User chooses a language, a theme or whether log groups start collapsed
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::account'}
- text: The Product keeps the choice in the browser and applies it at once
  kind: product
  actor: user
  entities:
  - {entity: account, effect: changes, facts: [Language, Theme, Collapse log groups]}
  contexts:
    web: {place: 'web::signed-in::account'}
---

# Change account settings

## Trigger

A person wants Woodpecker in another language or theme, or prefers logs collapsed.

## Outcome

The web UI uses the new settings in this browser.
