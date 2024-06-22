# Testing

## Backend

### Unit Tests

TODO

### Integration Tests

#### Pipeline Engine

The pipeline engine has a special backend called **`mock`** witch does not exec but emulate how a typical backend should behave.

an example pipeline confilg would be:

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

witch could be executed via `woodpecker-cli --log-level trace exec --backend-engine mock example.yaml`:

```none
...
9:18PM DBG pipeline/pipeline.go:94 > executing 2 stages, in order of: CLI=exec
9:18PM DBG pipeline/pipeline.go:104 > stage CLI=exec StagePos=0 Steps=echo
9:18PM DBG pipeline/pipeline.go:104 > stage CLI=exec StagePos=1 Steps=echo
9:18PM TRC pipeline/backend/mock/mock.go:75 > create workflow environment taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM DBG pipeline/pipeline.go:176 > prepare CLI=exec step=echo
9:18PM DBG pipeline/pipeline.go:203 > executing CLI=exec step=echo
9:18PM TRC pipeline/backend/mock/mock.go:81 > start step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM TRC pipeline/backend/mock/mock.go:167 > tail logs of step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM DBG pipeline/pipeline.go:209 > complete CLI=exec step=echo
[echo:L0:0s] StepName: echo
[echo:L1:0s] StepType: service
[echo:L2:0s] StepUUID: 01J10P578JQE6E25VV1A2DNQN9StepCommands:
[echo:L3:0s] 
[echo:L4:0s] echo ja
[echo:L5:0s] 9:18PM DBG pipeline/pipeline.go:176 > prepare CLI=exec step=echo
9:18PM DBG pipeline/pipeline.go:203 > executing CLI=exec step=echo
9:18PM TRC pipeline/backend/mock/mock.go:81 > start step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM TRC pipeline/backend/mock/mock.go:167 > tail logs of step echo taskUUID=01J10P578JQE6E25VV1EQF0745
[echo:L0:0s] StepName: echo
[echo:L1:0s] StepType: commands
[echo:L2:0s] StepUUID: 01J10P578JQE6E25VV1DFSXX1YStepCommands:
[echo:L3:0s] 
[echo:L4:0s] echo ja
[echo:L5:0s] 9:18PM TRC pipeline/backend/mock/mock.go:108 > wait for step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM TRC pipeline/backend/mock/mock.go:189 > stop step echo taskUUID=01J10P578JQE6E25VV1EQF0745
9:18PM DBG pipeline/pipeline.go:209 > complete CLI=exec step=echo
9:18PM TRC pipeline/backend/mock/mock.go:210 > delete workflow environment taskUUID=01J10P578JQE6E25VV1EQF0745
```

You can control the step behavior via its name:

- If you name it `step_start_fail` the engine will simulate a step start who fail (e.g. happens when the container image can not be pulled).
- If you name it `step_exec_error` the engine will simulate a command who executes with status code **1**.

There are also environment variables to alter things:

- `SLEEP` witch will simulate a given time duration as command execution time.
- `EXPECT_TYPE` witch let the backend error if set and the step is not the expected step-type.
