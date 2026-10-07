---
kind: primary
routes: {web: Web, api: API}
steps:
- text: The User sets the extension endpoints and chooses Save
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::extensions'}
    api: {place: 'api::signed-in'}
- text: The Product saves the Repository's extensions configuration
  kind: product
  actor: user
  entities:
  - entity: repository
    effect: changes
    facts: [Config extension endpoint, Config extension exclusive, Registry extension endpoint, Secret extension endpoint, Include netrc credentials]
  contexts:
    web: {place: 'web::signed-in::extensions'}
    api: {place: 'api::signed-in'}
---

# Change extensions settings

## Trigger

A team generates configuration, secrets or registries from its own service.

## Outcome

Woodpecker asks those extensions when it starts the repository's pipelines.
