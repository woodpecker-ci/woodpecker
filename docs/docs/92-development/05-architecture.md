# Architecture

## Package architecture

![Woodpecker architecture](./woodpecker-architecture.png)

## System architecture

### main package hirarchie

| package    | meaning                                                      | import
|------------|--------------------------------------------------------------|----------
| `cmd/**`   | parse commanline args & environment to stat server/cli/agent | all other
| `agent/**` | code only agent (remote worker) will need                    | `pipeline`, `shared`
| `cli/**`   | code only cli tool does need                                 | `pipeline`, `shared`, `woodpecker-go`
| `server/**`| code only server will need                                   | `pipeline`, `shared`
| `shared/**`| code shared for all three main tools (go help utils)         | only std and external libs
| `woodpecker-go/**` | go client for server rest api                        | std

### Server

| package            | meaning                                         | import
|--------------------|-------------------------------------------------|----------
| `server/api/**`    | handle web requests from `server/router`        | `pipeline`, `server/(badges\|ccmenue\|logging\|model\|pubsub\|queue\|remote\|shared\|store)`, `shared`, (TODO: mv `server/router/middleware/session`)
| `server/badges/**` | generate svg badges for pipelines               | `server/model`
| `server/ccmenu/**` | generate xml ccmenu for pipelines               | `server/model`
| `server/grpc/**`   | gRPC server agents can connect to               | `pipeline/rpc/**`, `server/(logging\|model\|pubsub\|queue\|remote\|shared\|store)`
| `server/logging/**`| logging lib for server... noop (TODO: rm)       |
| `server/model/**`  | structs for store (db) and api (json)           | std
| `server/plugins/**`| plugins for server                              | `server/model`, `server/remote`
| `server/pubsub/**` | pubsub lib for server (TODO: write down what exatly is handled there) | std
| `server/queue/**`  | queue lib for server (TODO: write down what exatly is handled there) | `server/model`
| `server/remote/**` | remote lib for server to connect and handle forge specific stuff | `shared`, `server/model`
| `server/router/**` | handle REST API (and all middleware) and serve route `web` | `shared`, `server/(api\|model\|remote\|store\|web)`
| `server/store/**`  | handle database                                 | `server/model`
| `server/shared/**` | shared utils only server need (TODO: import indecate unrelated func) | `server/(model\|remote\|store\|plugins)`, (TODO: mv `pipeline`)
| `server/web/**`    | server SPA                                      | `shared`, (TODO: mv `server/router/middleware/session`)


### Agent

TODO

### CLI

TODO
