---
domain: pipelines
relations:
- entity: workflow
  verb: has
  cardinality: one-to-many
references:
- kind: code
  role: implementation
  target: server/model/pipeline.go#Pipeline
- kind: code
  role: implementation
  target: server/model/const.go#StatusValue
---

# Pipeline

One run of a repository's pipeline configuration for one event, numbered per repository. It is made of Workflows.

## Information kept

- **Number** — its sequence number within the repository
- **Event** — Push, Tag, Pull Request, Pull Request merged/closed, Pull Request metadata changed, Release, Deploy, Cron or Manual
- **Branch** — the branch it ran for
- **Commit** — the commit it ran for
- **Message** — the commit message, or the message given for a manual run
- **Author** — who caused the event
- **Created** — when it was created
- **Duration** — when it started and finished
- **Changed files** — the files the event changed
- **Additional pipeline variables** — variables given for a manual run, deployment, restart or cron
- **Deployment target** — the target environment of a deployment
- **Deployment task** — the deployment task
- **Errors** — parse errors, linter warnings and runtime errors
- **Configuration** — the configuration files it was built from
- **Reviewer** — who approved or declined it, and when
- **Cancel info** — who canceled it, or which newer pipeline superseded it
- **Restarted from** — the pipeline it restarts or deploys
- **Cron** — the cron that started it
- **Woodpecker version** — the version it was executed on
- **Metadata** — everything a local run needs to replay it: repository, commit, event and environment

## States

### Pending

Waiting in the queue for an agent.

### Blocked

Waiting for someone with push access to approve or decline it.

### Running

At least one workflow is running.

### Success

Every workflow finished successfully.

### Failure

A workflow failed.

### Killed

Canceled while running or while awaiting approval, superseded by a newer pipeline, or stopped by its timeout.

### Canceled

Canceled while every workflow was still pending.

### Declined

Declined instead of approved; nothing ran.

### Error

The configuration could not be loaded or parsed, or the Product failed to run it.
