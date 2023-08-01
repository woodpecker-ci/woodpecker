# Local backend

:::danger
The local backend will execute the pipelines on the local system without any isolation of any kind.
:::

:::note
This backend is still pretty new and can not be treated as stable. Its
implementation and configuration can change at any time. Binary releases of the
agent will be available with the release of the [1.0.0
milestone](https://github.com/woodpecker-ci/woodpecker/milestone/4), so for now
you must compile the agent by yourself, to get the local backend functionality.
<!-- TODO: remove the self-compile note after the release of the agent -->
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

## Configuration

### Server

Enable connection to the server from the outside of the docker environment by
exposing the port 9000:

```yaml
# docker-compose.yml for the server
version: '3'

services:
  woodpecker-server:
  [...]
    ports:
      - 9000:9000
      [...]
    environment:
      - [...]
```

### Agent

You can use the `.env` file to store environmental variables for configuration.
At the minimum you need the following information:

```ini
# .env for the agent
WOODPECKER_AGENT_SECRET=replace_with_your_server_secret
WOODPECKER_SERVER=replace_with_your_server_address:9000
```

## Running the agent

Start the agent from the directory with the `.env` file:

`woodpecker-agent`

:::note
When using the `local` backend, the
[plugin-git](https://github.com/woodpecker-ci/plugin-git) binary must be in
your `$PATH` for the default clone step to work. If not, you can still write a
manual clone step.
:::

## Further configuration

### Specify the shell to be used for a pipeline step

The `image` entry is used to specify the shell, such as Bash or Fish, that is
used to run the commands.


```yaml
# .woodpecker.yml

steps:
  build:
    image: bash
    commands:
      [...]
```

### Using labels to filter tasks

You can use the [agent configuration
options](../15-agent-config.md#woodpecker_filter_labels) and the
[pipeline syntax](../../20-usage/20-pipeline-syntax.md#labels) to only run certain
pipelines on certain agents. Example:

Define a `label` `type` with value `exec` for a particular agent:

```ini
# .env for the agent

WOODPECKER_FILTER_LABELS=type=exec
```

Then, use this `label` `type` with value `exec` in the pipeline definition, to
only run on this agent:

```yaml
# .woodpecker.yml

labels:
  type: exec

steps:
  [...]
```

