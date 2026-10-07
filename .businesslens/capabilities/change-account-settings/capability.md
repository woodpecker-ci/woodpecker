---
availability:
- {place: 'web::signed-in'}
references:
- kind: code
  role: implementation
  target: web/src/views/user/UserGeneral.vue
- kind: code
  role: implementation
  target: web/src/compositions/useI18n.ts
- kind: code
  role: implementation
  target: web/src/compositions/useTheme.ts
- kind: code
  role: implementation
  target: web/src/compositions/useUserConfig.ts
---

# Change account settings

A signed-in person chooses the web UI's language and theme and whether log groups of finished steps start collapsed. The settings are kept in that browser, not on the server; the language starts from the browser's and falls back to English.
