---
toc_max_heading_level: 3
---

# Local backend

:::danger
The local backend will execute the pipelines on the local system without any isolation of any kind.
:::

:::note
Currently we do not support services for this backend.
[Read more here](https://github.com/woodpecker-ci/woodpecker/issues/3095).
:::

Since the code runs directly in the same context as the agent (same user, same
filesystem), a malicious pipeline could be used to access the agent
configuration especially the `WOODPECKER_AGENT_SECRET` variable.

It is recommended to use this backend only for private setup where the code and
pipeline can be trusted. You shouldn't use it for a public facing CI where
anyone can submit code or add new repositories. You shouldn't execute the agent
as a privileged user (root).

The local backend will use a random directory in $TMPDIR to store the cloned
code and execute commands.

In order to use this backend, you need to download (or build) the
[binary](https://github.com/woodpecker-ci/woodpecker/releases/latest) of the
agent, configure it and run it on the host machine.

## Usage

To enable the local backend, add this to your configuration:

```ini
WOODPECKER_BACKEND=local
```

### Shell

The `image` entry is used to specify the shell, such as Bash or Fish, that is
used to run the commands.

```yaml title=".woodpecker.yaml"
steps:
  build:
    image: bash
    commands: [...]
```

### Plugins

Plugins are just executable binaries:

```yaml
steps:
  build:
    image: /usr/bin/tree
```

If no commands are provided, we treat them as plugins in the usual manner.
In the context of the local backend, plugins are simply executable binaries, which can be located using their name if they are listed in `$PATH`, or through an absolute path.

### Options

#### `WOODPECKER_BACKEND_LOCAL_TEMP_DIR`

> Default: default temp directory

Directory to create folders for workflows.
