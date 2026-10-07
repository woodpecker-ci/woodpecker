---
entities:
- {entity: account, shows: [Language, Theme, Collapse log groups], collects: [Language, Theme, Collapse log groups]}
entryPoints:
- web: /user
references:
- kind: code
  role: implementation
  target: web/src/views/user/UserGeneral.vue
---

# Account

The person's own settings in this browser: the language of the web UI, its theme, and whether log groups start collapsed.
