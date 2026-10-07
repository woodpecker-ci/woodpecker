---
domain: pipelines
relations:
- entity: step
  verb: has
  cardinality: one-to-many
references:
- kind: code
  role: implementation
  target: server/model/workflow.go#Workflow
---

# Workflow

One configuration file of a pipeline (or one matrix combination of it), run by a single agent.

## Information kept

- **Name** — the workflow name, from its file
- **Platform** — the platform it runs on
- **Agent** — the agent that ran it
- **Matrix variables** — the matrix combination it runs
- **Depends on** — workflows that must finish first
- **Error** — why it could not run

## States

### Pending

Waiting for an agent.

### Blocked

Its pipeline awaits approval.

### Running

An agent is running it.

### Success

Every step succeeded or was allowed to fail.

### Failure

A step failed.

### Killed

Stopped while running, by a cancel or by the repository's timeout.

### Skipped

Never run, because its pipeline was canceled before it started or a workflow it depends on did not end as its conditions require.

### Declined

Its pipeline was declined.
