---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: repositories
references:
- {kind: code, role: implementation, target: server/api/user.go#GetRepos}
- {kind: code, role: implementation, target: web/src/views/Repos.vue}
- {kind: code, role: implementation, target: cli/repo/repo_list.go}
---

# Browse repositories

A person sees the enabled repositories they have access to, most recently visited first, or the repositories of one organization, and searches them.
