---
domain: pipelines
references:
- kind: code
  role: implementation
  target: server/model/step.go#Step
---

# Step

One clone, service, plugin, commands or cache step of a workflow, with its log.

## Information kept

- **Name** — the step name
- **Type** — clone, service, plugin, commands or cache
- **Exit code** — how the step's process ended
- **Error** — why it failed to run
- **Duration** — when it started and finished
- **Log** — the lines the step printed

## States

### Pending

Not started yet.

### Blocked

Its pipeline awaits approval.

### Running

Running on the agent.

### Success

Exited with code 0.

### Failure

Exited with another code or could not run.

### Killed

Stopped while running, or never run because its workflow was skipped.

### Skipped

Not run because its conditions did not match or an earlier step failed.

### Canceled

Canceled before it started.

### Declined

Its pipeline was declined.
