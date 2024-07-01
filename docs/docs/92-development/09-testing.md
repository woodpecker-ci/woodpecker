# Testing

## Backend

### Unit Tests

[We use default golang unit tests](https://go.dev/doc/tutorial/add-a-test)
with [`"github.com/stretchr/testify/assert"`](https://pkg.go.dev/github.com/stretchr/testify@v1.9.0/assert) to simplify testing.

### Integration Tests

### Dummy backend

There is a special backend called **`dummy`** which does not execute any commands, but emulates how a typical backend should behave.
To enable it you need to build the agent or cli with the `test` build tag.

An example pipeline config would be:

```yaml
when:
  event: manual

steps:
  - name: echo
    image: dummy
    commands: echo "hello woodpecker"
    environment:
      SLEEP: '1s'

services:
  echo:
    image: dummy
    commands: echo "i am a sevice"
```

This could be executed via `woodpecker-cli --log-level trace exec --backend-engine dummy example.yaml`:

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

There are also environment variables to alter step behaviour:

- `SLEEP: 10` will let the step wait 10 seconds
- `EXPECT_TYPE` allows to check if a step is a `clone`, `service`, `plugin` or `commands`
- `STEP_START_FAIL: true` if set will simulate a step to fail before actually being started (e.g. happens when the container image can not be pulled)
- `STEP_TAIL_FAIL: true` if set will error when we simulate to read from stdout for logs
- `STEP_EXIT_CODE: 2` if set will be used as exit code, default is 0
- `STEP_OOM_KILLED: true` simulates a step being killed by memory constrains

You can let the setup of a whole workflow fail by setting it's UUID to `WorkflowSetupShouldFail`.
