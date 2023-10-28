# CLI

```docker run --rm woodpeckerci/woodpecker-cli:v0.15```

```bash
NAME:
   woodpecker-cli - command line utility

USAGE:
   woodpecker-cli [global options] command [command options] [arguments...]

VERSION:
   v0.15.x

COMMANDS:
   build      manage pipelines
   log        manage logs
   deploy     deploy code
   exec       execute a pipeline locally
   info       show information about the current user
   registry   manage registries
   secret     manage secrets
   repo       manage repositories
   user       manage users
   lint       lint a pipeline configuration file
   log-level  get the logging level of the server, or set it with [level]
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --token value, -t value   server auth token [$WOODPECKER_TOKEN]
   --server value, -s value  server address [$WOODPECKER_SERVER]
   --log-level value         set logging level [$WOODPECKER_LOG_LEVEL]
   --help, -h                show help (default: false)
   --version, -v             print the version (default: false)
```
