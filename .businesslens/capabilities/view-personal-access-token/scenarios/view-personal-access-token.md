---
kind: primary
routes:
  web: Web
  api: API
steps:
- text: The User opens CLI & API
  kind: actor
  actor: user
  entities: []
  contexts:
    web: {place: 'web::signed-in::cli-and-api'}
    api: {place: 'api::signed-in'}
- text: The Product shows the Account's personal access token
  kind: product
  actor: user
  entities:
  - {entity: account, effect: reads, facts: [Personal access token]}
  contexts:
    web: {place: 'web::signed-in::cli-and-api'}
    api: {place: 'api::signed-in'}
---

# View the personal access token

## Trigger

A person wants to use the CLI or the API.

## Outcome

The person has their token and examples of using it.
