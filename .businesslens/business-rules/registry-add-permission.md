---
appliesTo:
- type: entity
  id: registry
  effect: creates
permits:
- related:
  - {verb: keeps, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Push, is: true}
- related:
  - {verb: keeps, entity: organization}
  - {verb: grants, entity: organization-membership}
  - {verb: holds, entity: user}
  when:
  - {entity: organization-membership, fact: Role, is: Admin}
- related:
  - {verb: keeps, entity: organization}
  - {verb: owns, entity: account}
  - {verb: uses, entity: user}
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/api/registry.go
- kind: code
  role: implementation
  target: server/router/middleware/session/user.go#MustOrgMember
- kind: code
  role: implementation
  target: server/router/api.go
---

# Only those who manage registry credentials add them

Registry credentials of a repository are managed by people with push access to it, those of a team organization by its admins, those of a personal organization by its owner, and global ones only by administrators.
