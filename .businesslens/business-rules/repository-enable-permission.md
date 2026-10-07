---
appliesTo:
- type: entity
  id: repository
  effect: creates
permits:
- related:
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Admin, is: true}
references:
- kind: code
  role: implementation
  target: server/api/repo.go#PostRepo
---

# Only a forge admin of a repository enables it

Woodpecker enables a repository only for someone the forge reports as an admin of it, and only when its owner is allowed on the server.
