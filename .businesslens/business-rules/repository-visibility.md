---
appliesTo:
- type: entity
  id: repository
  effect: reads
  contexts:
  - {place: 'web::public'}
  - {place: 'web::signed-in'}
  - {place: 'cli::signed-in'}
  - {place: 'api::public'}
  - {place: 'api::signed-in'}
permits:
- actors:
  - visitor
  when:
  - {fact: Project visibility, is: Public}
- actors:
  - user
  when:
  - {fact: Project visibility, is-not: Private}
- related:
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Pull, is: true}
- actors:
  - administrator
- unattended: true
references:
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#SetPerm
- kind: code
  role: implementation
  target: server/router/middleware/session/repo.go#MustPull
---

# Only people with pull access see a private repository

A public repository and its pipelines are seen by everyone, an internal one by every signed-in person, and a private one only by people the forge gives pull access, and by administrators. Status badges and CCMenu feeds are the exception: they show a repository's latest pipeline state whatever its visibility.
