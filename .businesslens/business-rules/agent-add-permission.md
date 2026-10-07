---
appliesTo:
- {type: entity, id: agent, effect: creates}
permits:
- related:
  - {verb: has, entity: organization}
  - {verb: grants, entity: organization-membership}
  - {verb: holds, entity: user}
  when:
  - {entity: organization-membership, fact: Role, is: Admin}
  - {entity: server-settings, fact: User-registered agents, is: Allowed}
- related:
  - {verb: has, entity: organization}
  - {verb: owns, entity: account}
  - {verb: uses, entity: user}
  when:
  - {entity: server-settings, fact: User-registered agents, is: Allowed}
- actors: [administrator]
references:
- {kind: code, role: implementation, target: server/api/agent.go}
- {kind: code, role: implementation, target: server/router/api.go}
---

# Only administrators and organization admins add agents

Server agents are managed by administrators; agents of a team organization by its admins and agents of a personal organization by its owner, while the server allows user-registered agents; when the operator disables them, only administrators manage agents.
