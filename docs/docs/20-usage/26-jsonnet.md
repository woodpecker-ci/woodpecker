# Jsonnet

Instead of writing every workflow as a separate YAML file, you can define your pipeline configuration in [Jsonnet](https://jsonnet.org/), a data templating language that compiles to JSON. This is useful when many workflows share structure — for example a matrix of build variants that differ only in a few settings.

Woodpecker compiles files with a `.jsonnet` extension natively. They can be placed in the same locations as YAML configs: a single `.woodpecker.jsonnet` at the repository root, or `.jsonnet` files inside the `.woodpecker/` folder (or your custom config path). `woodpecker-cli exec` and `woodpecker-cli lint` support them as well.

## Single workflow

A Jsonnet config that evaluates to a single object defines one workflow, just like a YAML file:

```jsonnet title=".woodpecker.jsonnet"
{
  steps: [
    {
      name: 'build',
      image: 'debian:stable-slim',
      commands: ['echo building'],
    },
  ],
}
```

The workflow is named after the file. Alternatively, a top-level `name` field sets the workflow name explicitly; it is removed from the config before it is processed.

## Multiple workflows from one file

A Jsonnet config that evaluates to an array of objects defines one workflow per element. Each element must have a unique `name`:

```jsonnet title=".woodpecker.jsonnet"
local variant(name, tag) = {
  name: name,
  steps: [
    {
      name: 'build',
      image: 'golang:' + tag,
      commands: ['go build ./...'],
    },
  ],
};

[
  variant('go-stable', '1'),
  variant('go-latest', 'rc'),
  variant('lint', '1') { steps: [{ name: 'lint', image: 'golangci/golangci-lint', commands: ['golangci-lint run'] }] },
]
```

This behaves exactly like placing `go-stable.yaml`, `go-latest.yaml` and `lint.yaml` in the `.woodpecker/` folder — workflows run in parallel and can be ordered with [`depends_on`](./25-workflows.md#flow-control).

## Limitations

- Configs must be self-contained: `import`/`importstr`, external variables (`std.extVar`) and native functions are not available. The full [Jsonnet standard library](https://jsonnet.org/ref/stdlib.html) can be used.
- Linter messages refer to the compiled JSON of a workflow, not to the Jsonnet source. Compile errors do reference the original source location.
- The compiled output is limited to 1 MiB.
