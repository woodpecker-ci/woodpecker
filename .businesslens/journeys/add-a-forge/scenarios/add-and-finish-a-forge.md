---
kind: primary
result: achieved
routes:
  web: Web
steps:
- text: The Administrator chooses the forge type, enters its URL, OAuth client and options and chooses Add
  kind: actor
  actor: administrator
  capability: add-forge
  entities:
  - {entity: forge, effect: reads, facts: []}
  contexts:
    web: {place: 'web::administration::forge'}
- text: The Product creates the Forge and opens its page
  kind: product
  actor: administrator
  capability: add-forge
  entities:
  - {entity: forge, effect: creates, facts: [Forge type, URL, OAuth client ID, OAuth client secret, OAuth host, Skip SSL verification, Allowed organizations, Advanced options]}
  contexts:
    web: {place: 'web::administration::forge'}
- text: The Administrator changes the new Forge's settings and chooses Save
  kind: actor
  actor: administrator
  capability: edit-forge
  entities:
  - {entity: forge, effect: reads, facts: [URL]}
  contexts:
    web: {place: 'web::administration::forge'}
- text: The Product saves the Forge
  kind: product
  actor: administrator
  capability: edit-forge
  entities:
  - {entity: forge, effect: changes, facts: [OAuth host, Skip SSL verification, Allowed organizations, Advanced options]}
  contexts:
    web: {place: 'web::administration::forge'}
---

# Add a forge and finish its settings

## Trigger

An administrator connects a forge from the Forges page.

## Outcome

The forge is connected with the settings the administrator chose on its page.
