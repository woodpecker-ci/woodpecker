# Parameters

Parameters are typed inputs for manual pipeline runs. Instead of typing free-text key/value variables every time you trigger a pipeline manually, a repository can define a set of parameters that are rendered as proper input widgets (dropdown, checkbox, number or text field) in the manual run form, pre-filled with default values.

The chosen values are passed to the pipeline as environment variables, using the same mechanism as the "additional pipeline variables" of the manual run form. Parameters only apply to manual runs: pipelines triggered by other events (push, tag, pull request, cron, ...) are not affected.

To configure parameters you need at least push access to the repository.

## Define parameters

Parameters are managed in the repository settings under the **Parameters** tab.

Each parameter has:

- **Name**: the environment variable name the value is exposed as (e.g. `DEPLOY_TARGET`). Must be a valid environment variable identifier and unique per repository.
- **Type**: one of:
  - `string`: free-form text input
  - `number`: numeric input
  - `boolean`: checkbox, passed as `true` or `false`
  - `choice`: dropdown with a predefined list of options
- **Description**: shown below the parameter name in the manual run form.
- **Options**: the allowed values (`choice` type only).
- **Default value**: pre-filled in the run form. If a run is triggered without a value (e.g. via the API), the default is applied server-side.
- **Required**: the run cannot be started without a value. Not available for `boolean`, as a checkbox always has a value.
- **Position**: parameters are sorted ascending by this value in the run form.

## Trigger a manual run

When triggering a manual pipeline run, the defined parameters are shown as input widgets above the free-text "additional pipeline variables" editor. Submitted values are validated server-side: required parameters must be set and `choice` values must be one of the configured options.

Use the values in your workflow like any other environment variable:

```yaml
when:
  - event: manual

steps:
  - name: deploy
    image: alpine
    commands:
      - echo "deploying to $DEPLOY_TARGET"
```

:::note
Both parameters and ad-hoc additional variables end up in the same environment variable namespace. On a name collision, the parameter value wins.
:::

## Prefill the run form via URL

The manual run form can be pre-filled through URL query parameters, which is handy for bookmarks or linking from other tools:

```txt
https://ci.example.com/repos/42/manual?branch=main&DEPLOY_TARGET=production&SOME_VAR=foo
```

- `branch` selects the branch (if it exists).
- A key matching a defined parameter prefills its widget. Values that are invalid for the parameter type are ignored.
- Any other key prefills the additional variables editor.

The values are only pre-filled — the run still has to be reviewed and submitted.
