---
toc_max_heading_level: 2
---

# Local

:::danger
The local backend executes pipelines on the local system without any isolation.
:::

:::note
Currently we do not support [services](../../../20-usage/60-services.md) for this backend.
[Read more here](https://github.com/woodpecker-ci/woodpecker/issues/3095).
:::

Since the commands run directly in the same context as the agent (same user, same
filesystem), a malicious pipeline could be used to access the agent
configuration especially the `WOODPECKER_AGENT_SECRET` variable.

It is recommended to use this backend only for private setup where the code and
pipeline can be trusted. It should not be used in a public instance where
anyone can submit code or add new repositories. The agent should not run as a privileged user (root).

The local backend will use a random directory in `$TMPDIR` to store the cloned
code and execute commands.

In order to use this backend, you need to download (or build) the
[agent](https://github.com/woodpecker-ci/woodpecker/releases/latest), configure it and run it on the host machine.

## Step specific configuration

### Shell

The `image` entrypoint is used to specify the shell, such as `bash` or `fish`, that is
used to run the commands.

```yaml title=".woodpecker.yaml"
steps:
  - name: build
    image: bash
    commands: [...]
```

### Plugins

```yaml
steps:
  - name: build
    image: /usr/bin/tree
```

If no commands are provided, plugins are treated in the usual manner.
In the context of the local backend, plugins are simply executable binaries, which can be located using their name if they are listed in `$PATH`, or through an absolute path.

## Environment variables

### `WOODPECKER_BACKEND_LOCAL_TEMP_DIR`

> Default: default temp directory

Directory to create folders for workflows.
