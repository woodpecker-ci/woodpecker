# Testing

## Backend

### Unit Tests

[We use default golang unit tests.](https://go.dev/doc/tutorial/add-a-test)
With [`"github.com/stretchr/testify/assert"`](https://pkg.go.dev/github.com/stretchr/testify@v1.9.0/assert) to simplify the test code.

### Integration Tests

#### Pipeline Engine

The pipeline engine has a special backend called **`dummy`** witch does not exec but emulate how a typical backend should behave.

An example pipeline config would be:

```yaml
when:
  event: manual

steps:
  - name: echo
    image: dummy
    commands: echo ja
    environment:
      SLEEP: "1s"

services:
  echo:
    image: dummy
    commands: echo ja
```

witch could be executed via `woodpecker-cli --log-level trace exec --backend-engine dummy example.yaml`:

```none
9:18PM DBG pipeline/pipeline.go:94 > executing 2 stages, in order of: CLI=exec
9:18PM DBG pipeline/pipeline.go:104 > stage CLI=exec StagePos=0 Steps=echo
9:18PM DBG pipeline/pipeline.go:104 > stage CLI=exec StagePos=1 Steps=echo
9:18PM TRC pipeline/backend/dummy/dummy.go:75 > create workflow environment taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM DBG pipeline/pipeline.go:176 > prepare CLI=exec step=echo
9:18PM DBG pipeline/pipeline.go:203 > executing CLI=exec step=echo
9:18PM TRC pipeline/backend/dummy/dummy.go:81 > start step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM TRC pipeline/backend/dummy/dummy.go:167 > tail logs of step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM DBG pipeline/pipeline.go:209 > complete CLI=exec step=echo
[echo:L0:0s] StepName: echo
[echo:L1:0s] StepType: service
[echo:L2:0s] StepUUID: 01J10P578JQE6E25VV1A2DNQN9
[echo:L3:0s] StepCommands:
[echo:L4:0s] ------------------
[echo:L5:0s] echo ja
[echo:L6:0s] ------------------
[echo:L7:0s] 9:18PM DBG pipeline/pipeline.go:176 > prepare CLI=exec step=echo
9:18PM DBG pipeline/pipeline.go:203 > executing CLI=exec step=echo
9:18PM TRC pipeline/backend/dummy/dummy.go:81 > start step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM TRC pipeline/backend/dummy/dummy.go:167 > tail logs of step echo taskUUID=01J10P578JQE6E25VV1EQF0745
[echo:L0:0s] StepName: echo
[echo:L1:0s] StepType: commands
[echo:L2:0s] StepUUID: 01J10P578JQE6E25VV1DFSXX1Y
[echo:L3:0s] StepCommands:
[echo:L4:0s] ------------------
[echo:L5:0s] echo ja
[echo:L6:0s] ------------------
[echo:L7:0s] 9:18PM TRC pipeline/backend/dummy/dummy.go:108 > wait for step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM TRC pipeline/backend/dummy/dummy.go:187 > stop step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM DBG pipeline/pipeline.go:209 > complete CLI=exec step=echo
9:18PM TRC pipeline/backend/dummy/dummy.go:208 > delete workflow environment taskUUID=01J10P578JQE6E25VV1EQF0745
```

You can control the step behavior via its name:

- If you name it `step_start_fail` the engine will simulate a step start who fail (e.g. happens when the container image can not be pulled).
- If you name it `step_exec_error` the engine will simulate a command who executes with status code **1**.

There are also environment variables to alter things:

- `SLEEP` witch will simulate a given time duration as command execution time.
- `EXPECT_TYPE` witch let the backend error if set and the step is not the expected step-type.
