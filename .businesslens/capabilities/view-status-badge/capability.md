---
availability:
- {place: 'api::public'}
- {place: 'web::signed-in'}
references:
- kind: code
  role: implementation
  target: server/api/badge.go#GetBadge
- kind: code
  role: implementation
  target: server/api/badge.go#GetCC
- kind: code
  role: implementation
  target: web/src/views/repo/settings/Badge.vue
- kind: doc
  role: context
  target: docs/docs/20-usage/80-badges.md
---

# View status badge

Anyone fetches a badge image showing the state of a repository's latest pipeline on a branch (the default branch unless one is named) for given events (push unless named), or of one workflow or step of it, or a CCMenu feed of it. Badges are served whatever the repository's visibility. A repository admin builds the badge's address on the repository's Badge settings.
