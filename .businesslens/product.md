---
id: woodpecker
summary: A self-hosted CI/CD engine that runs pipelines defined in a repository whenever its forge reports a change, on agents the operator runs.
category: continuous-integration
tags:
- ci
- cd
- pipelines
- self-hosted
authors:
- name: Woodpecker Authors
  url: https://woodpecker-ci.org
license: Apache-2.0
languages:
- ar
- bar
- bn
- cs
- de
- en
- eo
- es
- fi
- fr
- hi
- hu
- id
- it
- ja
- ko
- lo
- lv
- nb-NO
- nl
- pl
- pt
- ru
- th
- uk
- ur
- zh-Hans
- zh-Hant
limitations:
- People sign in with their account on a forge; Woodpecker keeps no passwords of its own.
- Repository permissions come from the forge; Woodpecker does not let anyone grant pull, push or admin access itself.
- Pipelines are defined in YAML files kept in the repository, or returned by a configured config extension.
- Pipelines run only on agents that the operator starts and connects to the server.
- Server-wide settings such as open registration, the admin list, allowed organizations and default project settings are set by the operator when the server starts; only the log level changes while it runs.
- Secret and registry values are never shown again once saved.
- Each workflow runs in its own workspace; workflows pass files to each other only through what their steps do.
- A pipeline run locally with the CLI runs its workflows one after another and receives no secrets or registries from a server.
- Deleting or disabling a repository in Woodpecker leaves the repository on the forge untouched.
references:
- kind: doc
  role: context
  target: README.md
- kind: doc
  role: context
  target: docs/docs/20-usage/15-terminology/index.md
- kind: doc
  role: context
  target: https://woodpecker-ci.org/docs/intro
---

# Woodpecker

Woodpecker runs continuous integration and delivery pipelines for repositories
hosted on a forge (GitHub, GitLab, Gitea, Forgejo, Bitbucket, Bitbucket Data
Center, or an addon). When someone enables a repository, Woodpecker installs a
webhook on the forge; every push, tag, release, pull request or deployment the
forge reports, every cron schedule, and every manual run becomes a Pipeline of
Workflows and Steps that agents execute. People follow pipelines and their logs
in the web UI, the CLI or the API, approve pipelines from untrusted sources, and keep the
secrets, registry credentials and crons their pipelines need.

## Intent

Give teams a simple, yet powerful CI/CD engine with great extensibility that
they host themselves next to the forge they already use.
