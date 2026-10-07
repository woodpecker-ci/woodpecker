---
availability:
- {place: 'web::signed-in'}
- {place: 'api::signed-in'}
references:
- {kind: code, role: implementation, target: web/src/views/repo/settings/Extensions.vue}
- {kind: code, role: implementation, target: server/services/config/http.go}
- {kind: doc, role: context, target: docs/docs/20-usage/72-extensions/index.md}
---

# Change extensions settings

A repository admin points the repository at HTTP extensions that supply its pipeline configuration, registries or secrets, and chooses whether forge credentials are sent to them. The page shows the public key extensions use to verify Woodpecker's calls.
