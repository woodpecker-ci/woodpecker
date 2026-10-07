---
relations:
- entity: organization
  verb: owns
  cardinality: one-to-one
references:
- kind: code
  role: implementation
  target: server/model/user.go#User
---

# Account

The record Woodpecker keeps for a person who signs in through a forge, whichever role they hold.

## Information kept

- **Login** — the person's username on the forge
- **Email** — the email address the forge reports
- **Avatar URL** — the picture shown next to the person
- **Admin** — whether the person is a server administrator, set by an administrator or by the server's admin list; a login on the admin list is made an administrator again at each sign-in
- **Personal access token** — the token the CLI and API accept on the person's behalf; resetting it also ends every session of the person
- **Language** — the language of the web UI, kept in the person's browser
- **Theme** — Auto, Light or Dark, kept in the person's browser
- **Collapse log groups** — whether a finished step's log opens with its command groups collapsed, kept in the person's browser
