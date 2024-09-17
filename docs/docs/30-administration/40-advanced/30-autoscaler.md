# Autoscaler

If your would like dynamically scale your agents with the load, you can use [our autoscaler](https://github.com/woodpecker-ci/autoscaler).

Please note that the autoscaler is not feature-complete yet. You can follow the progress [here](https://github.com/woodpecker-ci/autoscaler#roadmap).

## Setup

### docker compose

If you are using docker compose you can add the following to your `docker-compose.yaml` file:

```yaml
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:next
    [...]

  woodpecker-autoscaler:
    image: woodpeckerci/autoscaler:next
    restart: always
    depends_on:
      - woodpecker-server
    environment:
      - WOODPECKER_SERVER=https://your-woodpecker-server.tld # the url of your woodpecker server / could also be a public url
      - WOODPECKER_TOKEN=${WOODPECKER_TOKEN} # the api token you can get from the UI https://your-woodpecker-server.tld/user
      - WOODPECKER_MIN_AGENTS=0
      - WOODPECKER_MAX_AGENTS=3
      - WOODPECKER_WORKFLOWS_PER_AGENT=2 # the number of workflows each agent can run at the same time
      - WOODPECKER_GRPC_ADDR=https://grpc.your-woodpecker-server.tld # the grpc address of your woodpecker server, publicly accessible from the agents
      - WOODPECKER_GRPC_SECURE=true
      - WOODPECKER_AGENT_ENV= # optional environment variables to pass to the agents
      - WOODPECKER_PROVIDER=hetznercloud # set the provider, you can find all the available ones down below
      - WOODPECKER_HETZNERCLOUD_API_TOKEN=${WOODPECKER_HETZNERCLOUD_API_TOKEN} # your api token for the Hetzner cloud
```
