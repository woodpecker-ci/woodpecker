---
availability:
- {place: 'cli::local'}
- {place: 'web::signed-in'}
references:
- kind: code
  role: implementation
  target: cli/setup/setup.go
- kind: code
  role: implementation
  target: cli/setup/token_fetcher.go
- kind: code
  role: implementation
  target: web/src/views/cli/Auth.vue
---

# Sign in to CLI

Setting the CLI up for a server: the CLI opens the server's web UI, the signed-in person confirms, and the CLI keeps the server address and the person's personal access token as a context.
