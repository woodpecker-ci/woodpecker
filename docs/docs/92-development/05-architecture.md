# Architecture

## Module Interactions

![Woodpecker architecture](./woodpecker-architecture.svg)

<!--
  To update the graph, first look at a simple svg of all module imports:
  `go run github.com/loov/goda@latest graph 'go.woodpecker-ci.org/woodpecker/v3/...' | dot -Tsvg -o graph.svg`

  generate a new svg of the graph using:
  `dot -Tsvg woodpecker-architecture.dot -o woodpecker-architecture.svg`
-->

## System architecture

### main package hierarchy

| package            | meaning                                                        | imports                               |
| ------------------ | -------------------------------------------------------------- | ------------------------------------- |
| `cmd/**`           | parse command-line args & environment to stat server/cli/agent | all other                             |
| `agent/**`         | code only agent (remote worker) will need                      | `pipeline`, `rpc`, `shared`           |
| `cli/**`           | code only cli tool does need                                   | `pipeline`, `shared`, `woodpecker-go` |
| `server/**`        | code only server will need                                     | `pipeline`, `rpc`, `shared`           |
| `pipeline/**`      | core ci/cd engine from parsing to execution                    | `shared`                              |
| `rpc/**`           | RPC interface for agent-server communication                   | `pipeline`                            |
| `shared/**`        | code shared for all three main tools (go help utils)           | only std and external libs            |
| `woodpecker-go/**` | go client for server rest api                                  | std                                   |

### Server

| package              | meaning                                                                        | imports                                                                                                                                                                                      |
| -------------------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `server/api/**`      | handle web requests from `server/router`                                       | `pipeline`, `rpc`, `../badges`, `../ccmenu`, `../logging`, `../model`, `../pubsub`, `../queue`, `../forge`, `../shared`, `../store`, `shared`, (TODO: mv `server/router/middleware/session`) |
| `server/badges/**`   | generate svg badges for pipelines                                              | `../model`                                                                                                                                                                                   |
| `server/ccmenu/**`   | generate xml ccmenu for pipelines                                              | `../model`                                                                                                                                                                                   |
| `server/rpc/**`      | gRPC server agents can connect to                                              | `rpc`, `../logging`, `../model`, `../pubsub`, `../queue`, `../forge`, `../pipeline`, `../store`                                                                                              |
| `server/logging/**`  | logging lib for gPRC server to stream logs while running                       | std                                                                                                                                                                                          |
| `server/model/**`    | structs for store (db) and api (json)                                          | std                                                                                                                                                                                          |
| `server/pipeline/**` | orchestrate pipelines (TODO: parts of it should move into /pipeline)           | `pipeline`, `../model`, `../pubsub`, `../queue`, `../forge`, `../store`, `../plugins`                                                                                                        |
| `server/pubsub/**`   | pubsub lib for server to push changes to the WebUI                             | std                                                                                                                                                                                          |
| `server/queue/**`    | queue lib for server where agents pull new pipelines from via gRPC             | `server/model`                                                                                                                                                                               |
| `server/forge/**`    | forge lib for server to connect and handle forge specific stuff                | `shared`, `server/model`                                                                                                                                                                     |
| `server/router/**`   | handle requests to REST API (and all middleware) and serve UI and WebUI config | `shared`, `../api`, `../model`, `../forge`, `../store`, `../web`                                                                                                                             |
| `server/store/**`    | handle database                                                                | `server/model`                                                                                                                                                                               |
| `server/web/**`      | server SPA                                                                     |                                                                                                                                                                                              |

- `../` = `server/`

### Agent

| package        | meaning                                              | imports                                                |
| -------------- | ---------------------------------------------------- | ------------------------------------------------------ |
| `agent/**`     | agent implementation that runs workflows             | `pipeline`, `rpc`, `shared`                            |
| `agent/rpc/**` | gRPC client for agent-server communication           | `rpc`, `pipeline/backend/types`, std and external libs |
| `cmd/agent/**` | CLI interface for starting and configuring the agent | `agent`, std and external libs                         |

The agent is a remote worker that connects to the server via gRPC to receive pipeline execution instructions and report back execution state and logs.
The agent polls the server's queue for new work, executes pipeline steps using the pipeline engine, and streams results back to the server.

TODO: Review cmd/agent/core to determine if any logic should be moved into the agent package for better separation of concerns.

### CLI

| package                  | meaning                                                                 | imports                                                                          |
| ------------------------ | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `cli/admin/**`           | admin commands for server management (users, secrets, registries, etc.) | `../common`, `../internal`, `woodpecker-go`                                      |
| `cli/common/**`          | shared utilities and helpers used across all CLI subcommands            | `../internal/config`, `../update`, `shared`                                      |
| `cli/context/**`         | manage multiple server contexts (connections to different servers)      | `../common`, `../internal/config`, `../output`                                   |
| `cli/exec/**`            | execute pipelines locally without server orchestration                  | `pipeline`, `../common`, `../lint`, `shared`                                     |
| `cli/info/**`            | display information about the current user                              | `../common`, `../internal`                                                       |
| `cli/internal/**`        | internal utilities for HTTP client, auth, and server communication      | `../internal/config`, `woodpecker-go`, `shared`                                  |
| `cli/internal/config/**` | configuration file management (load, store, credentials)                | std and external libs                                                            |
| `cli/lint/**`            | validate pipeline configuration files                                   | `pipeline/frontend/yaml`, `pipeline/frontend/yaml/linter`, `../common`, `shared` |
| `cli/org/**`             | manage organization-level resources (secrets, registries)               | `../common`, `../internal`, `woodpecker-go`                                      |
| `cli/output/**`          | formatting utilities for CLI output (tables, etc.)                      | std and external libs                                                            |
| `cli/pipeline/**`        | manage pipeline operations (start, stop, approve, logs, etc.)           | `../common`, `../internal`, `../output`, `woodpecker-go`, `shared`               |
| `cli/repo/**`            | manage repository-level resources (repos, crons, secrets, registries)   | `../common`, `../internal`, `../output`, `woodpecker-go`                         |
| `cli/setup/**`           | interactive first-time setup wizard for CLI configuration               | `../internal/config`                                                             |
| `cli/update/**`          | self-updater for the CLI binary                                         | std and external libs                                                            |
| `cmd/cli/**`             | CLI entry point and command structure                                   | `cli/**`                                                                         |

The CLI provides a command-line interface for interacting with Woodpecker servers.
Each subcommand is organized into its own package under `cli/<subcommand>/`.

The `cli/exec` subcommand allows local pipeline execution for testing and development by combining pipeline parsing and execution without requiring a running server or agent.

- `../` = `cli/`

### Engine

The engine is the shared kernel that validates, parses frontend facing config files, enrich it by the provided forge metadata and produce config for the backends to execute on based on that. It also contains the default backend implementations.

#### Runtime

The runtime is the package controlling how a workflow is executed, and can be found at `pipeline/runtime`.

A workflow is one call of `Run(runnerCtx)`. Its stages run one after the other, the steps of a stage run in parallel.
The workflow context `r.ctx` is canceled if the workflow is canceled or timed out, `runnerCtx` outlives it so the cleanup can still reach the backend.

```mermaid
stateDiagram-v2
    direction TB

    [*] --> ValidateConfig: Run(runnerCtx)
    ValidateConfig --> [*]: tracer, logger or spec is nil,<br/>return error
    ValidateConfig --> SetupWorkflow: defer DestroyWorkflow()
    SetupWorkflow --> DestroyWorkflow: error, trace setup error
    SetupWorkflow --> Stage: ok

    state "Stage: runStage(steps)" as Stage {
        [*] --> StepA: executeStep(step)
        StepA --> [*]
        --
        [*] --> StepB: executeStep(step)
        StepB --> [*]
        --
        [*] --> DetachedStep: executeStep(step)
        DetachedStep --> [*]: started, return nil,<br/>pipeline continues
    }

    Stage --> Stage: errgroup collected the errors,<br/>r.err.Set(), next stage
    Stage --> DestroyWorkflow: all stages done
    Stage --> DestroyWorkflow: r.ctx canceled,<br/>wait for the running stage
    DestroyWorkflow --> [*]: setup error or canceled,<br/>return error
    DestroyWorkflow --> WaitDetached: all stages done
    WaitDetached --> [*]: return nil,<br/>or the error if it is no step failure

    note right of ValidateConfig
        Checks values a user has no control over
    end note
    note right of SetupWorkflow
        Blocks: the workflow
        Calls: SetupWorkflow(r.ctx)
    end note
    note right of Stage
        All steps in parallel (errgroup),
        one goroutine per step
        Blocks: the next stage, until every
        step goroutine returned
        A detached step returns once it is started
    end note
    note right of DestroyWorkflow
        Runs exactly once, also on every early return
        Calls: DestroyWorkflow(runnerCtx)
        Uses a 5s shutdown context if runnerCtx is done
    end note
    note right of WaitDetached
        Blocks: the return of Run()
        On: detached steps completing in the
        background, with their logs and traces
    end note
```

Each step of a stage goes through `executeStep(step)`.
It is the only place where the runtime calls the step functions of the backend:

```mermaid
stateDiagram-v2
    direction TB

    [*] --> Skipped: shouldSkipStep(),<br/>OnSuccess / OnFailure check
    Skipped --> [*]: trace skip, return nil

    [*] --> Starting: traceStep(nil, nil, step),<br/>emit "step started" trace
    Starting --> Tailing: StartStep ok
    Starting --> Failed: StartStep error

    Tailing --> Running: TailStep ok,<br/>log goroutine started
    Tailing --> Failed: TailStep error

    Running --> Exiting: log stream drained
    Exiting --> Cleanup: WaitStep returned the state
    Exiting --> Failed: WaitStep error
    Cleanup --> Done: DestroyStep ok
    Cleanup --> Failed: DestroyStep error

    Failed --> [*]: traceStep(nil, err, step),<br/>return err
    Done --> [*]: traceStep(state, err, step),<br/>emit "step completed" trace,<br/>return err (exit code, OOM, canceled)

    note right of Starting
        setStepEnv(): CI_* variables of the step,
        if plugin: SetDroneEnviron()
        Blocks: this step goroutine
        On: the backend starting the step, e.g. image pull
        Calls: StartStep(r.ctx)
    end note
    note right of Tailing
        Calls: TailStep(r.ctx), returns the log stream
        startStep() = StartStep + TailStep + log goroutine
    end note
    note right of Running
        step.Detached? yes: executeStep() returns nil here,
        the rest runs in a background goroutine and its
        errors are logged, not returned
        no: blocks this step goroutine
        On: the backend closing the log stream
    end note
    note right of Exiting
        completeStep() = drain logs, Wait, Destroy
        Blocks: this step goroutine
        On: the step process to exit
        Calls: WaitStep(r.ctx)
    end note
    note right of Cleanup
        Calls: DestroyStep(runnerCtx),
        so it still works after a cancel
    end note
    note left of Done
        FailureIgnore? The error of a blocking
        step is suppressed if failure is ignored
    end note
```
