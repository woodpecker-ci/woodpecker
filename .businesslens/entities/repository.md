---
relations:
- entity: pipeline
  verb: has
  cardinality: one-to-many
- entity: secret
  verb: keeps
  cardinality: one-to-many
- entity: registry
  verb: keeps
  cardinality: one-to-many
- entity: cron
  verb: schedules
  cardinality: one-to-many
- entity: repository-permission
  verb: grants
  cardinality: one-to-many
references:
- kind: code
  role: implementation
  target: server/model/repo.go#Repo
---

# Repository

A forge repository that has been enabled in Woodpecker, with the project settings that decide how its pipelines are started and run.

## Information kept

- **Full name** — owner and name, as on the forge
- **Forge link** — where the repository is opened on the forge
- **Default branch** — the branch crons and badges use when none is named
- **Project visibility** — Public, Private or Internal; private repositories on the forge start Private, others Public
- **Pipeline path** — where the pipeline configuration is read from; empty means the default paths
- **Allow pull requests** — whether pull request events start pipelines
- **Allow deployments** — whether pipelines may be deployed; off when a repository is first enabled
- **Approval requirements** — which events wait for approval: none, pull requests from forks, all pull requests or all events from the forge
- **Allowed users** — people whose events never wait for approval
- **Timeout** — how many minutes a workflow may run
- **Cancel previous pipelines** — the events for which a new pipeline cancels the still active earlier ones of the same branch or ref
- **Trusted** — whether pipelines may use network, volume and security-sensitive options
- **Custom trusted clone plugins** — clone plugins that receive the forge credentials
- **Config extension endpoint** — an HTTP service asked for the pipeline configuration
- **Config extension exclusive** — whether the config extension replaces every other way of fetching configuration
- **Registry extension endpoint** — an HTTP service asked for registry credentials
- **Secret extension endpoint** — an HTTP service asked for secrets
- **Include netrc credentials** — whether the forge credentials are sent to each extension
- **Owner** — the person whose forge credentials Woodpecker uses for the repository
- **Webhook** — the hook Woodpecker installed on the forge to receive events

## States

### Enabled

Woodpecker receives the repository's events and runs its pipelines.

### Disabled

The webhook is removed and no new pipelines start; its pipelines, secrets and settings are kept.
