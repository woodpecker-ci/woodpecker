---
kind: system
acts: external
relations:
- entity: repository
  verb: hosts
  cardinality: one-to-many
- entity: organization
  verb: hosts
  cardinality: one-to-many
- entity: account
  verb: hosts
  cardinality: one-to-many
references:
- kind: code
  role: implementation
  target: server/model/forge.go#Forge
- kind: code
  role: implementation
  target: server/forge/forge.go
---

# Forge

A code hosting service Woodpecker is connected to. People sign in through it, its repositories are what Woodpecker builds, and it sends Woodpecker webhooks for the events that start pipelines.

## Information kept

- **Forge type** — GitHub, GitLab, Gitea, Forgejo, Bitbucket, Bitbucket Data Center or Addon
- **URL** — where the forge is reached
- **OAuth client ID** — the OAuth application Woodpecker signs people in with
- **OAuth client secret** — the secret of that OAuth application, never shown again once saved
- **OAuth host** — the public URL for OAuth when it differs from the URL
- **Skip SSL verification** — whether Woodpecker skips certificate checks when calling the forge
- **Allowed organizations** — organizations whose members may sign in through this forge, in addition to the server-wide list
- **Advanced options** — forge-type specific options such as merge ref, public only, Git username and password, or the addon executable
