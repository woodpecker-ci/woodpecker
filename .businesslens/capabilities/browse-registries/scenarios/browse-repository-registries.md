---
kind: primary
routes: {web: Web, cli: CLI, api: API}
steps:
- text: The User opens the registries of a Repository
  kind: actor
  actor: user
  entities:
  - entity: repository
    effect: reads
    facts: [Full name]
  contexts:
    web: {place: 'web::signed-in::repository-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
- text: The Product lists the Repository's Registries with those of its organization and the server
  kind: product
  actor: user
  entities:
  - entity: registry
    effect: reads
    facts: [Address, Username, Read only]
  - entity: organization
    effect: reads
    facts: []
  - entity: repository
    effect: reads
    facts: []
  contexts:
    web: {place: 'web::signed-in::repository-registries'}
    cli: {place: 'cli::signed-in'}
    api: {place: 'api::signed-in'}
---

# Browse a repository's registries

## Trigger

A person checks which registries a repository's pipelines receive.

## Outcome

The person sees the registries without their confidential values.
