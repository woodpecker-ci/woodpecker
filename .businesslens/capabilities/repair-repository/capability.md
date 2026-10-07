---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: repository-actions
references:
- {kind: code, role: implementation, target: server/api/repo.go#RepairRepo}
- {kind: code, role: implementation, target: cli/repo/repo_repair.go}
---

# Repair repository

A repository admin has Woodpecker install the repository's webhook again and refresh its name and details from the forge.
