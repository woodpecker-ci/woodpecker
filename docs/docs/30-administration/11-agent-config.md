# Agent configuration

Agents are configured by the command line or environement variables. At the minimum you need the following information:

```yaml
# docker-compose.yml
version: '3'

services:
  woodpecker-agent:
  [...]
  environment:
+   - WOODPECKER_SERVER=localhost:9000
+   - WOODPECKER_AGENT_SECRET="your-shared-secret-goes-here"

```

## Processes per agent

By default the maximum processes that are run per agent is 1. If required you can add `WOODPECKER_MAX_PROCS` to increase your parellel processing on a per-agent basis.

```yaml
# docker-compose.yml
version: '3'

services:
  woodpecker-agent:
  [...]
  environment:
    - WOODPECKER_SERVER=localhost:9000
    - WOODPECKER_AGENT_SECRET="your-shared-secret-goes-here"
+    - WOODPECKER_MAX_PROCS=4
```

## Filtering agents

When building your pipelines as long as you have set the platform or filter, builds can be made to only run code on certain agents. 

```
- WOODPECKER_HOSTNAME=mycompany-ci-01.example.com
- WOODPECKER_PLATFORM=linux/amd64
- WOODPECKER_FILTER=???
```

### Filter on Platform

Only want certain pipelines or steps to run on certain platforms? Such as arm vs amd64? 

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-agent:
  [...]
  environment:
    - WOODPECKER_SERVER=localhost:9000
    - WOODPECKER_AGENT_SECRET=""
+   - WOODPECKER_PLATFORM=linux/arm64
```

```yaml
# .woodpecker.yml
pipeline:
  build:
   image: golang
   commands:
     - go build
     - go test
  when:
    platform: linux/amd64


  testing:
   image: golang
   commands:
     - go build
     - go test
  when:
    platform: linux/arm*


```

See [Conditionals Pipeline](usage/pipeline-syntax#step-when---conditional-execution) syntax for more


## All agent configuration options

Here is the full list of configuration options and their default variables. 

```yaml
    - WOODPECKER_SERVER=localhost:9000
    - WOODPECKER_AGENT_SECRET=""
    - WOODPECKER_USERNAME=x-oauth-basic
    - WOODPECKER_DEBUG=true
    - WOODPECKER_LOG_LEVEL=""
    - WOODPECKER_DEBUG_PRETTY=""
    - WOODPECKER_DEBUG_NOCOLOR=true
    - WOODPECKER_HOSTNAME=""
    - WOODPECKER_PLATFORM="linux/amd64"
    - WOODPECKER_FILTER=""
    - WOODPECKER_MAX_PROCS=1
    - WOODPECKER_HEALTHCHECK=true
    - WOODPECKER_KEEPALIVE_TIME=10
    - WOODPECKER_KEEPALIVE_TIMEOUT=time.Second * 20
    - WOODPECKER_GRPC_SECURE=""
    - WOODPECKER_GRPC_VERIFY=true
```
