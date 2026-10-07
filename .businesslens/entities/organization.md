---
relations:
- entity: repository
  verb: owns
  cardinality: one-to-many
- entity: secret
  verb: keeps
  cardinality: one-to-many
- entity: registry
  verb: keeps
  cardinality: one-to-many
- entity: agent
  verb: has
  cardinality: one-to-many
- entity: organization-membership
  verb: grants
  cardinality: one-to-many
references:
- kind: code
  role: implementation
  target: server/model/org.go#Org
- kind: code
  role: implementation
  target: server/api/org.go
---

# Organization

An owner namespace on a forge, such as a team or a person, whose repositories Woodpecker builds. Every person who signs in also has a personal organization named after their login, which holds their own secrets, registries and agents.

## Information kept

- **Name** — the owner name on the forge
- **Personal** — whether it is a person's own namespace rather than a team
