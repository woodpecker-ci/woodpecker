---
toc_max_heading_level: 3
---

# Compile

A compile workflow generates pipeline configuration. It runs on an agent before
any ordinary workflow starts, and what it prints becomes configuration for the
rest of the same pipeline.

This is for pipelines whose shape depends on something Woodpecker cannot know in
advance: a monorepo that should only build the packages that changed, a matrix
whose axes come from a file in the repository, or a set of workflows generated
from a manifest.

:::note
A compile workflow runs code from the repository on an agent, exactly like any
other workflow, and is subject to the same trust, secret and approval rules. A
pipeline that requires approval does not run its compile workflows until a human
approves it.
:::

## Declaring compile steps

Add a `compile` section to any workflow config. Its steps look like ordinary
steps:

```yaml
when:
  event: push

compile:
  - name: generate
    image: alpine
    commands:
      - ./scripts/generate-pipeline.sh

steps:
  - name: test
    image: golang
    commands:
      - go test ./...
```

A config may declare `compile`, `steps`, or both. A config with only a `compile`
section contributes nothing to the run phase by itself.

The compile workflow inherits the config's `when`, `labels`, `clone`,
`workspace`, `concurrency` and timeout, so it is routed to an agent the same way
the ordinary workflow would be. Services are not started for it: a config
generator does not need sidecars.

`depends_on` inside `compile` may only name another compile step, and
`depends_on` inside `steps` only another ordinary step. The two sections become
separate workflows. Detached compile steps are rejected, because a step that
keeps writing after the workflow has moved on cannot have its output captured
reliably.

## The two phases

A pipeline that declares compile steps runs in two phases:

1. Every compile workflow runs. Nothing else is created yet.
2. The ordinary workflows run, built from the configuration phase 1 produced.

Dependencies are resolved within a phase. A compile workflow and an ordinary
workflow may share a name, and neither can `depends_on` the other.

There are exactly two phases. A configuration emitted by a compile workflow may
not itself declare `compile`.

## Emitting configuration

A compile step emits configuration by printing a block to its standard output:

<!-- cspell:disable -->

```text
-----BEGIN WOODPECKER CONFIG RESPONSE-----
eyJ2ZXJzaW9uIjoxLCJjb25maWdzIjpbey...
-----END WOODPECKER CONFIG RESPONSE-----
```

<!-- cspell:enable -->

The payload is base64 of JSON, wrapped at 64 columns. The JSON is:

```json
{
  "version": 1,
  "configs": [
    { "name": "build", "data": "steps:\n  - name: test\n    image: golang\n" },
    { "name": "deploy", "data": "" }
  ]
}
```

Each entry either replaces the configuration of that name, adds a new one, or,
when `data` is empty, removes it.

This is the same shape as the response of an
[external configuration service](./72-extensions/40-configuration-extension.md),
so an existing configuration service can become a compile plugin by changing how
it is invoked rather than what it returns.

Most backends merge a step's standard output and standard error into one stream,
so a block can be broken up by anything else the step prints at the same moment.
The emitting step should print its response and nothing else.

### What counts as a valid response

| The step prints                  | Result                       |
| -------------------------------- | ---------------------------- |
| a block with configs             | those configs are merged     |
| a block with `"configs": []`     | the pipeline runs unchanged  |
| no block at all                  | the pipeline fails           |
| a malformed or unterminated block| the pipeline fails           |

Emitting nothing is an error rather than "no change", so a generator that
crashes before it prints anything cannot be mistaken for one that deliberately
decided to change nothing.

The block stays in the step's log for auditing, but is hidden from the ordinary
log view.

### Limits

- at most 64 configurations per pipeline
- at most 4 MiB of configuration in total
- an emitted name must be a plain file name, with no path separators

Removing a configuration that is not part of the pipeline is an error. A removal
that silently does nothing means an extra workflow runs, and the pipeline stays
green while doing more than the configuration asked for.

### Secrets

A compile workflow gets the ordinary secret set of the repository. A generator
that needs a secret in the pipeline it produces should emit `from_secret` rather
than the value: an embedded value ends up in a stored configuration.

## What the pipeline is built from

The configurations in the repository, and anything an external configuration
service returned, are the **source** configurations. What a compile workflow
emits is merged on top of them.

Generated configurations are recorded against the pipeline and can be inspected
afterwards, but they are never used as input again. Restarting a pipeline
rebuilds from the source configurations and runs the compile phase again, so a
restart always compiles again.

That means a restart is reproducible in its *input*, not necessarily in its
*output*: a generator that reads the network or the clock can produce a
different pipeline the second time. The recorded configuration makes that
visible when it happens.

## Local execution

`woodpecker-cli exec` runs the compile phase too. The compile workflows execute
first, their output is merged, and the resulting workflows run afterwards, so a
generator can be tested without pushing.
