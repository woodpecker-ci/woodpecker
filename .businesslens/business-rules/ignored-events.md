---
appliesTo:
- type: capability
  id: trigger-pipeline
- type: entity
  id: repository
  facts:
  - Allow pull requests
references:
- kind: code
  role: implementation
  target: server/api/hook.go#PostHook
---

# Disabled repositories and pull requests a repository does not allow start no pipeline

Events of a disabled repository, or of one without an owner, are ignored, and so are pull request events while the repository does not allow pull requests.
