---
availability:
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
references:
- {kind: code, role: implementation, target: server/api/repo.go#ChownRepo}
- {kind: code, role: implementation, target: cli/repo/repo_chown.go}
---

# Assume repository ownership

A repository admin makes themselves the repository's owner, so Woodpecker uses their forge credentials for it.
