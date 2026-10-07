---
kind: person
acts: external
relations:
- entity: account
  verb: uses
  cardinality: one-to-one
- entity: repository-permission
  verb: holds
  cardinality: one-to-many
- entity: organization-membership
  verb: holds
  cardinality: one-to-many
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustUser
---

# User

A signed-in person who enables repositories they administer on the forge and works with the pipelines, secrets, registries, crons and agents of repositories and organizations they have access to.
