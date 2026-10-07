---
entities:
- {entity: forge, shows: [Forge type, URL]}
entryPoints:
- web: /login
references:
- kind: code
  role: implementation
  target: web/src/views/Login.vue
---

# Sign in

Where a visitor chooses the forge to sign in with, and sees why a sign-in was refused.
