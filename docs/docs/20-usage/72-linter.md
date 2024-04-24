# Linter

Woodpecker automatically lints your workflow files for errors, deprecations and bad habits. Errors and warnings are shown in the UI for any pipelines.

![errors and warnings in UI](./linter-warnings-errors.png)

## Running the linter from CLI

You can run the linter also manually from the CLI:

```shell
woodpecker-cli lint <workflow files>
```

## Bad habit warnings

Woodpecker warns you if your configuration contains some bad habits.

### Event filter for all steps

All your items in `when` blocks should have an `event` filter, so no step runs on all events. This is recommended because if new events are added, your steps probably shouldn't run.

Example:

```yaml
when:
  - branch: main
  - event: push
```

This will trigger the warning because the first item (`branch: main`) does not filter with an event.
