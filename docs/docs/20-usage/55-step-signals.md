# Step signals

A step can tell Woodpecker what to do next by exiting with a specific exit code. You map exit codes to **signals** under the `signals` option of a step. When the step exits with a mapped code, Woodpecker performs the signal's action instead of treating the code as a failure.

```yaml
steps:
  - name: check-affected
    image: node
    commands:
      - npx nx show projects --affected --type app | grep -qx my-app || exit 78
    signals:
      skip: 78

  - name: build
    image: woodpeckerci/plugin-docker-buildx
    settings:
      repo: my-org/my-app
```

If `my-app` is not affected, `check-affected` exits with `78`: it is shown as successful, `build` is shown as skipped and the workflow stays successful.

Signals are opt-in. A step without `signals` behaves as before: every exit code other than `0` is a failure.

## Signals

| Signal    | The step is shown as | Effect                                                                                                                         |
| --------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `skip`    | successful           | The steps that would run after it are skipped. The workflow stays successful.                                                  |
| `success` | successful           | Nothing else, the workflow continues as usual.                                                                                 |
| `restart` | (runs again)         | The step runs again, up to `max_restarts` times. If the last attempt still exits with a `restart` code, the step fails.        |
| `cancel`  | successful           | The workflow is canceled: running steps are stopped, the remaining steps do not run and the workflow is shown as killed.       |
| `pause`   | successful           | The workflow is paused until a user resumes or declines it in the UI, or until an optional timeout.                            |

## Syntax

Each signal takes an exit code or a list of exit codes:

```yaml
signals:
  skip: 78
  success: [1, 2]
  restart: 75
  cancel: 70
  pause: 71
```

Each signal also accepts a long form with `exit_codes` and, depending on the signal, options:

```yaml
signals:
  restart:
    exit_codes: [75, 76]
    max_restarts: 5 # restart only, see below
  pause:
    exit_codes: 71
    timeout: 30m # pause only, see below
    on_timeout: resume # pause only, see below
```

Rules, checked by the [linter](./72-linter.md):

- Exit codes must be between `1` and `255`. `0` always means success.
- An exit code can only be mapped to one signal of a step.
- Unknown signals are rejected.

### Choosing exit codes

Avoid `126`, `127` and codes of `128` and above: shells use them for "command cannot be executed", "command not found" and "killed by signal N" (for example `137` when a container runs out of memory). The linter warns about them. Codes between `64` and `113` are a safe choice. `78` is the code Drone used for skipping, so it is a good default for `skip`.

A step that is killed because it ran out of memory always fails, even if `137` is mapped to a signal.

## `skip`

The step is shown as successful and the steps that would start after it are shown as skipped, including steps with `when: status: [success, failure]` (for example notification or cleanup steps). The workflow stays successful.

Steps that run at the same time as the skipping step (steps that do not depend on it and start in parallel with it) still run to the end, and their result still counts: if one of them fails, the workflow fails. Only the steps that would start after them are skipped.

A workflow that already failed before the skip stays failed.

## `success`

The step is shown as successful and the workflow continues as usual. Use it for tools that use non-zero exit codes for results that should not fail the pipeline, for example a linter that exits with `1` when it only found warnings:

```yaml
steps:
  - name: lint
    image: golangci/golangci-lint
    commands:
      - ./lint.sh # exits 1 on warnings, 2 on errors
    signals:
      success: 1
```

## `restart`

The step runs again. Use it for steps that fail because of temporary problems, for example network errors:

```yaml
steps:
  - name: deploy
    image: alpine
    commands:
      - ./deploy.sh || exit 75 # 75 = temporary failure, try again
    signals:
      restart: 75
```

By default the step is restarted up to 2 times (3 attempts in total). Use `max_restarts` (between `1` and `10`) to change it:

```yaml
signals:
  restart:
    exit_codes: [75, 76]
    max_restarts: 5 # up to 6 attempts in total
```

- Each attempt starts from a new container with the same workspace, so files written by an earlier attempt are still there.
- A line like `--- woodpecker: step sent restart signal, starting attempt 2 of 3 ---` in the step log separates the attempts.
- If the last attempt still exits with a `restart` code, the step fails with that code.
- If an attempt exits with a code mapped to another signal, that signal is applied.
- Detached steps and services are not restarted: for them the `restart` codes count as failure.

## `cancel`

The step is shown as successful and its workflow is canceled: steps running at the same time are stopped, the remaining steps do not run and the workflow is shown as killed.

Only the workflow of the step is canceled. Other workflows of the pipeline, for example other [matrix](./30-matrix-workflows.md) entries, keep running.

## `pause`

The step is shown as successful and the workflow pauses once the steps running at the same time have finished. The workflow is shown as paused, with **Resume** and **Decline** buttons. Users with push access to the repository can:

- **Resume** the workflow: the remaining steps run.
- **Decline** the workflow: the remaining steps do not run and the workflow is shown as killed.

```yaml
steps:
  - name: build
    image: alpine
    commands:
      - ./build.sh

  - name: approve-deploy
    image: alpine
    commands:
      - exit 71
    signals:
      pause: 71

  - name: deploy
    image: alpine
    commands:
      - ./deploy.sh
```

### Timeout

By default a paused workflow waits until it is resumed, declined or reaches the workflow [timeout](./75-project-settings.md#timeout). Set `timeout` to decide automatically after a while, and `on_timeout` to choose what happens then:

```yaml
signals:
  pause:
    exit_codes: 71
    timeout: 30m # a duration like 90s, 30m, 1h or 1h30m
    on_timeout: resume # or decline (default)
```

- `decline` (default): the workflow is declined, as if a user clicked **Decline**.
- `resume`: the workflow continues, as if a user clicked **Resume**. Use it for "wait for a veto": the workflow goes on unless somebody declines it in time.

The log of the step that paused the workflow shows how long the workflow waits and what happened, for example:

```text
--- woodpecker: workflow paused, waiting up to 30m0s for resume or decline, then: resume ---
--- woodpecker: no decision within 30m0s, workflow resumed ---
```

The timeout counts from the moment the workflow pauses and is kept when the agent reconnects to the server. A `timeout` longer than the workflow timeout has no effect: the workflow timeout ends the workflow first.

### API

The same can be done through the API:

- `POST /api/repos/{repo_id}/pipelines/{pipeline_number}/workflows/{workflow_id}/resume`
- `POST /api/repos/{repo_id}/pipelines/{pipeline_number}/workflows/{workflow_id}/decline`

Keep in mind:

- A paused workflow keeps its agent busy, and the waiting time counts towards the workflow [timeout](./75-project-settings.md#timeout).
- Only the workflow of the step pauses. Other workflows of the pipeline keep running.
- Canceling the pipeline also cancels a paused workflow.
- If the step is in the last stage of the workflow, there is nothing left to wait for, so the workflow does not pause.
- A decision is kept in the server's memory until the agent picks it up. If the server restarts right after you resumed or declined and the workflow is still paused, resume or decline it again.

## Matrix workflows

Signals only affect the workflow of the step. In a [matrix](./30-matrix-workflows.md), every entry is its own workflow: one entry can skip, cancel or pause without changing the others. This makes it possible to start one matrix entry per app and let each entry decide at runtime whether it has to build:

```yaml
matrix:
  APP:
    - app-a
    - app-b

steps:
  - name: check-affected
    image: node
    commands:
      - npx nx show projects --affected --type app > affected.txt
      - grep -qx "${APP}" affected.txt || exit 78
    signals:
      skip: 78

  - name: build
    image: woodpeckerci/plugin-docker-buildx
    settings:
      repo: my-org/${APP}
```

## Local execution

[`woodpecker-cli exec`](./73-local-execution.md) supports `skip`, `success`, `restart` and `cancel`. It does not support `pause`: it logs a warning and continues with the next steps.

## Requirements

The server reads `signals` from the configuration and the agent applies them, so both need a Woodpecker version that supports step signals. Agents and servers must use the same version.
