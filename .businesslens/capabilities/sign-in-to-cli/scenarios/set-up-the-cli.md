---
kind: primary
routes:
  setup: Setup
steps:
- text: The Visitor runs setup with the server address
  kind: actor
  actor: visitor
  entities: []
  contexts:
    setup: {place: 'cli::local'}
- text: The Product opens the server's Sign in to CLI page in the browser
  kind: product
  actor: visitor
  entities: []
  contexts:
    setup: {place: 'cli::local'}
- text: The User confirms signing in to the CLI
  kind: actor
  actor: user
  entities: []
  contexts:
    setup: {place: 'web::signed-in::cli-sign-in'}
- text: The Product hands the Account's personal access token to the waiting CLI
  kind: product
  actor: user
  entities:
  - {entity: account, effect: reads, facts: [Personal access token]}
  contexts:
    setup: {place: 'web::signed-in::cli-sign-in'}
- text: The Product saves the server and token as the CLI's context
  kind: product
  actor: visitor
  entities: []
  contexts:
    setup: {place: 'cli::local'}
---

# Set up the CLI with a server

## Trigger

A person runs `woodpecker-cli setup`.

## Outcome

The CLI is signed in to the server and the browser tab tells the person to return to the CLI.

## Edge cases

- The person chooses Abort: the CLI is told the sign-in was denied and saves nothing.
- The person is not signed in to the web UI: signing in comes first, then the confirmation page.
- A token can also be given directly to setup, skipping the browser.
