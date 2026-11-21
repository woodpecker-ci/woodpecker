# Step signals

A step can tell Woodpecker what to do next by exiting with a specific exit code. You map exit codes to **signals** under the `signals` option of a step. When the step exits with a mapped code, Woodpecker performs the signal's action instead of treating the code as a failure.

```yaml
steps:
  - name: check-affected
    image: node
    commands:
      - affected=$$(npx nx show projects --affected --type app) || exit 1
      - rc=0; printf '%s\n' "$$affected" | grep -Fxq my-app || rc=$$?
      - if [ "$$rc" -eq 1 ]; then exit 78; fi
      - exit "$$rc"
    signals:
      skip: 78

  - name: build
    image: woodpeckerci/plugin-docker-buildx
    settings:
      repo: my-org/my-app
```

If `my-app` is not affected, `check-affected` exits with `78`: it is shown as successful, `build` is shown as skipped and the workflow stays successful. If `nx` or `grep` fails, the step fails.

:::warning
Only exit with a signal's code when you are sure of the result. A shortcut like `nx ... | grep -q my-app || exit 78` also exits with `78` when `nx` itself fails, which turns an error into a green skip.
:::

Signals are opt-in. A step without `signals` behaves as before: every exit code other than `0` is a failure.

## Signals

| Signal    | The step is shown as | Effect                                                                                                                   |
| --------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `skip`    | successful           | The steps that would run after it are skipped. The workflow stays successful.                                            |
| `success` | successful           | Nothing else, the workflow continues as usual.                                                                           |
| `restart` | (runs again)         | The step runs again, up to `max_restarts` times. If the last attempt still exits with a `restart` code, the step fails.  |
| `cancel`  | successful           | The workflow is canceled: running steps are stopped, the remaining steps do not run and the workflow is shown as killed. |

## Syntax

Each signal takes an exit code or a list of exit codes:

```yaml
signals:
  skip: 78
  success: [1, 2]
  restart: 75
  cancel: 70
```

Each signal also accepts a long form with `exit_codes` and, depending on the signal, options:

```yaml
signals:
  restart:
    exit_codes: [75, 76]
    max_restarts: 5 # restart only, see below
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

The skip takes precedence over `when: status`. If a step running in parallel fails, the workflow ends as failed, but a following step with `when: status: [failure]` is still skipped. Do not rely on failure notifications or cleanup steps after a step that can send `skip`.

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

## Matrix workflows

Signals only affect the workflow of the step. In a [matrix](./30-matrix-workflows.md), every entry is its own workflow: one entry can skip or cancel without changing the others. This makes it possible to start one matrix entry per app and let each entry decide at runtime whether it has to build:

```yaml
matrix:
  APP:
    - app-a
    - app-b

steps:
  - name: check-affected
    image: node
    commands:
      - affected=$$(npx nx show projects --affected --type app) || exit 1
      - rc=0; printf '%s\n' "$$affected" | grep -Fxq "${APP}" || rc=$$?
      - if [ "$$rc" -eq 1 ]; then exit 78; fi
      - exit "$$rc"
    signals:
      skip: 78

  - name: build
    image: woodpeckerci/plugin-docker-buildx
    settings:
      repo: my-org/${APP}
```

## Local execution

[`woodpecker-cli exec`](./73-local-execution.md) supports all signals.

## Requirements

The server reads `signals` from the configuration and the agent applies them, so both need a Woodpecker version that supports step signals. Agents and servers must use the same version.
