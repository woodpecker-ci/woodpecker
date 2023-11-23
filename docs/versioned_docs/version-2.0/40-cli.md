# CLI

# NAME

woodpecker-cli - command line utility

# SYNOPSIS

woodpecker-cli

```
[--log-file]=[value]
[--log-level]=[value]
[--nocolor]
[--pretty]
[--server|-s]=[value]
[--token|-t]=[value]
```

**Usage**:

```
woodpecker-cli [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token


# COMMANDS

## pipeline

manage pipelines

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

show pipeline history

**--branch**="": branch filter

**--event**="": event filter

**--format**="": format output (default: [33mPipeline #{{ .Number }} [0m
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
)

**--limit**="": limit the list size (default: 0)

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--status**="": status filter

**--token, -t**="": server auth token

### last

show latest pipeline details

**--branch**="": branch name (default: main)

**--format**="": format output (default: Number: {{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}
)

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### logs

show pipeline logs

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show pipeline details

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### stop

stop a pipeline

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### start

start a pipeline

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### approve

approve a pipeline

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### decline

decline a pipeline

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### queue

show pipeline queue

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### ps

show pipeline steps

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### create

create new pipeline

**--branch**="": branch to create pipeline from

**--format**="": format output (default: [33mPipeline #{{ .Number }} [0m
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
)

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

**--var**="": key=value

## log

manage logs

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### purge

purge a log

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

## deploy

deploy code

**--branch**="": branch filter (default: main)

**--event**="": event filter (default: push)

**--format**="": format output (default: Number: {{ .Number }}
Status: {{ .Status }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}
Target: {{ .Deploy }}
)

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--param, -p**="": custom parameters to be injected into the step environment. Format: KEY=value

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--status**="": status filter (default: success)

**--token, -t**="": server auth token

## exec

execute a local pipeline

**--backend-docker-api-version**="": the version of the API to reach, leave empty for latest.

**--backend-docker-cert**="": path to load the TLS certificates for connecting to docker server

**--backend-docker-host**="": path to docker socket or url to the docker server

**--backend-docker-ipv6**: backend docker enable IPV6

**--backend-docker-network**="": backend docker network

**--backend-docker-tls-verify**: enable or disable TLS verification for connecting to docker server

**--backend-docker-volumes**="": backend docker volumes (comma separated)

**--backend-engine**="": backend engine to run pipelines on (default: auto-detect)

**--backend-http-proxy**="": if set, pass the environment variable down as "HTTP_PROXY" to steps

**--backend-https-proxy**="": if set, pass the environment variable down as "HTTPS_PROXY" to steps

**--backend-k8s-namespace**="": backend k8s namespace (default: woodpecker)

**--backend-k8s-pod-annotations**="": backend k8s additional worker pod annotations

**--backend-k8s-pod-labels**="": backend k8s additional worker pod labels

**--backend-k8s-storage-class**="": backend k8s storage class

**--backend-k8s-storage-rwx**: backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)

**--backend-k8s-volume-size**="": backend k8s volume size (default 10G) (default: 10G)

**--backend-local-temp-dir**="": set a different temp dir to clone workflows into (default: /tmp)

**--backend-no-proxy**="": if set, pass the environment variable down as "NO_PROXY" to steps

**--commit-author-avatar**="": 

**--commit-author-email**="": 

**--commit-author-name**="": 

**--commit-branch**="": 

**--commit-message**="": 

**--commit-ref**="": 

**--commit-refspec**="": 

**--commit-sha**="": 

**--connect-retry-count**="": number of times to retry connecting to the server (default: 0)

**--connect-retry-delay**="": duration to wait before retrying to connect to the server (default: 0s)

**--env**="": 

**--forge-type**="": 

**--forge-url**="": 

**--local**: run from local directory

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--netrc-machine**="": 

**--netrc-password**="": 

**--netrc-username**="": 

**--network**="": external networks

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pipeline-created**="":  (default: 0)

**--pipeline-event**="":  (default: manual)

**--pipeline-finished**="":  (default: 0)

**--pipeline-number**="":  (default: 0)

**--pipeline-parent**="":  (default: 0)

**--pipeline-started**="":  (default: 0)

**--pipeline-status**="": 

**--pipeline-target**="": 

**--pipeline-url**="": 

**--pretty**: enable pretty-printed debug output

**--prev-commit-author-avatar**="": 

**--prev-commit-author-email**="": 

**--prev-commit-author-name**="": 

**--prev-commit-branch**="": 

**--prev-commit-message**="": 

**--prev-commit-ref**="": 

**--prev-commit-refspec**="": 

**--prev-commit-sha**="": 

**--prev-pipeline-created**="":  (default: 0)

**--prev-pipeline-event**="": 

**--prev-pipeline-finished**="":  (default: 0)

**--prev-pipeline-number**="":  (default: 0)

**--prev-pipeline-started**="":  (default: 0)

**--prev-pipeline-status**="": 

**--prev-pipeline-url**="": 

**--privileged**="": privileged plugins (default: "plugins/docker", "plugins/gcr", "plugins/ecr", "woodpeckerci/plugin-docker-buildx", "codeberg.org/woodpecker-plugins/docker-buildx")

**--repo**="": full repo name

**--repo-clone-ssh-url**="": 

**--repo-clone-url**="": 

**--repo-private**="": 

**--repo-remote-id**="": 

**--repo-trusted**: 

**--repo-url**="": 

**--server, -s**="": server address

**--step-name**="":  (default: 0)

**--system-name**="":  (default: woodpecker)

**--system-platform**="": 

**--system-url**="":  (default: https://github.com/woodpecker-ci/woodpecker)

**--timeout**="": pipeline timeout (default: 0s)

**--token, -t**="": server auth token

**--volumes**="": pipeline volumes

**--workflow-name**="":  (default: 0)

**--workflow-number**="":  (default: 0)

**--workspace-base**="":  (default: /woodpecker)

**--workspace-path**="":  (default: src)

## info

show information about the current user

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

## registry

manage registries

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

adds a registry

**--hostname**="": registry hostname (default: docker.io)

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--password**="": registry password

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--username**="": registry username

### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--password**="": registry password

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--username**="": registry username

### info

display registry info

**--hostname**="": registry hostname (default: docker.io)

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list registries

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

## secret

manage secrets

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

adds a secret

**--events**="": secret limited to these events

**--global**: global secret

**--images**="": secret limited to these images

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--value**="": secret value

### rm

remove a secret

**--global**: global secret

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a secret

**--events**="": secret limited to these events

**--global**: global secret

**--images**="": secret limited to these images

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--value**="": secret value

### info

display secret info

**--global**: global secret

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list secrets

**--global**: global secret

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

## repo

manage repositories

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list all repos

**--format**="": format output (default: [33m{{ .FullName }}[0m (id: {{ .ID }}, forgeRemoteID: {{ .ForgeRemoteID }}))

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--org**="": filter by organization

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show repository details

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

add a repository

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a repository

**--config**="": repository configuration path (e.g. .woodpecker.yml)

**--gated**: repository is gated

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pipeline-counter**="": repository starting pipeline number (default: 0)

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--timeout**="": repository timeout (default: 0s)

**--token, -t**="": server auth token

**--trusted**: repository is trusted

**--unsafe**: validate updating the pipeline-counter is unsafe

**--visibility**="": repository visibility

### rm

remove a repository

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### repair

repair repository webhooks

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### chown

assume ownership of a repository

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### sync

synchronize the repository list

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

## user

manage users

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list all users

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show user details

**--format**="": format output (default: {{ .Login }})

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

adds a user

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### rm

remove a user

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

## lint

lint a pipeline configuration file

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

## log-level

get the logging level of the server, or set it with [level]

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

## cron

manage cron jobs

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

add a cron job

**--branch**="": cron branch

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--name**="": cron name

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

**--server, -s**="": server address

**--token, -t**="": server auth token

### rm

remove a cron job

**--id**="": cron id

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a cron job

**--branch**="": cron branch

**--id**="": cron id

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--name**="": cron name

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

display info about a cron job

**--id**="": cron id

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list cron jobs

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token
