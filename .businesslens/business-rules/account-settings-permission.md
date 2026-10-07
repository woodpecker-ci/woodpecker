---
appliesTo:
- type: entity
  id: account
  effect: changes
  facts:
  - Language
  - Theme
  - Collapse log groups
permits:
- related:
  - {verb: uses, entity: user}
- related:
  - {verb: uses, entity: administrator}
references:
- kind: code
  role: implementation
  target: web/src/views/user/UserGeneral.vue
---

# Only the person changes their own account settings

A person's language, theme and log settings are kept in their own browser, so only they change them.
