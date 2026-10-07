---
type: webhook
actors:
- forge
references:
- kind: code
  role: implementation
  target: server/api/hook.go#PostHook
---

# Forge webhook

The webhook Woodpecker installs on the forge for every enabled repository. The forge calls it for pushes, tags, releases, pull requests and deployments.
