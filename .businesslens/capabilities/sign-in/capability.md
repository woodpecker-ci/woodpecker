---
availability:
- {place: 'web::public'}
references:
- kind: code
  role: implementation
  target: server/api/login.go#HandleAuth
- kind: code
  role: implementation
  target: web/src/views/Login.vue
---

# Sign in

A visitor signs in through one of the server's forges with OAuth. The first sign-in registers an account and a personal organization while registration is open or the person is on the admin list; organizations the server or the forge restricts sign-in to are enforced on every sign-in.
