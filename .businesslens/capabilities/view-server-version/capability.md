---
availability:
- {place: 'web::administration'}
- {place: 'api::public'}
references:
- kind: code
  role: implementation
  target: web/src/views/admin/AdminInfo.vue
- kind: code
  role: implementation
  target: web/src/compositions/useVersion.ts
- kind: code
  role: implementation
  target: server/api/z.go#Version
---

# View server version

Anyone asks the API which Woodpecker version the server runs; administrators see it on the admin Info page, with a notice when a newer release is available unless the version check is turned off.
