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

**--log-file**="": where logs are written to. 'stdout' and 'stderr' can be used as special keywords (default: "stderr")

**--log-level**="": set logging level (default: "info")

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token


# COMMANDS

## pipeline

manage pipelines

### ls

show pipeline history

**--branch**="": branch filter

**--event**="": event filter

**--format**="": format output (default: "\x1b[33mPipeline #{{ .Number }} \x1b[0m\nStatus: {{ .Status }}\nEvent: {{ .Event }}\nCommit: {{ .Commit }}\nBranch: {{ .Branch }}\nRef: {{ .Ref }}\nAuthor: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}\nMessage: {{ .Message }}\n")

**--limit**="": limit the list size (default: 25)

**--status**="": status filter

### last

show latest pipeline details

**--branch**="": branch name (default: "main")

**--format**="": format output (default: "Number: {{ .Number }}\nStatus: {{ .Status }}\nEvent: {{ .Event }}\nCommit: {{ .Commit }}\nBranch: {{ .Branch }}\nRef: {{ .Ref }}\nMessage: {{ .Message }}\nAuthor: {{ .Author }}\n")

### logs

show pipeline logs

### info

show pipeline details

**--format**="": format output (default: "Number: {{ .Number }}\nStatus: {{ .Status }}\nEvent: {{ .Event }}\nCommit: {{ .Commit }}\nBranch: {{ .Branch }}\nRef: {{ .Ref }}\nMessage: {{ .Message }}\nAuthor: {{ .Author }}\n")

### stop

stop a pipeline

### start

start a pipeline

**--param, -p**="": custom parameters to be injected into the step environment. Format: KEY=value

### approve

approve a pipeline

### decline

decline a pipeline

### queue

show pipeline queue

**--format**="": format output (default: "\x1b[33m{{ .FullName }} #{{ .Number }} \x1b[0m\nStatus: {{ .Status }}\nEvent: {{ .Event }}\nCommit: {{ .Commit }}\nBranch: {{ .Branch }}\nRef: {{ .Ref }}\nAuthor: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}\nMessage: {{ .Message }}\n")

### ps

show pipeline steps

**--format**="": format output (default: "\x1b[33mStep #{{ .PID }} \x1b[0m\nStep: {{ .Name }}\nState: {{ .State }}\n")

### create

create new pipeline

**--branch**="": branch to create pipeline from

**--format**="": format output (default: "\x1b[33mPipeline #{{ .Number }} \x1b[0m\nStatus: {{ .Status }}\nEvent: {{ .Event }}\nCommit: {{ .Commit }}\nBranch: {{ .Branch }}\nRef: {{ .Ref }}\nAuthor: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}\nMessage: {{ .Message }}\n")

**--var**="": key=value

## log

manage logs

### purge

purge a log

## deploy

deploy code

**--branch**="": branch filter (default: "main")

**--event**="": event filter (default: "push")

**--format**="": format output (default: "Number: {{ .Number }}\nStatus: {{ .Status }}\nCommit: {{ .Commit }}\nBranch: {{ .Branch }}\nRef: {{ .Ref }}\nMessage: {{ .Message }}\nAuthor: {{ .Author }}\nTarget: {{ .Deploy }}\n")

**--param, -p**="": custom parameters to be injected into the step environment. Format: KEY=value

**--status**="": status filter (default: "success")

## exec

execute a local pipeline

**--backend-docker-api-version**="": the version of the API to reach, leave empty for latest.

**--backend-docker-cert**="": path to load the TLS certificates for connecting to docker server

**--backend-docker-host**="": path to docker socket or url to the docker server

**--backend-docker-ipv6**: backend docker enable IPV6

**--backend-docker-network**="": backend docker network

**--backend-docker-tls-verify**: enable or disable TLS verification for connecting to docker server

**--backend-docker-volumes**="": backend docker volumes (comma separated)

**--backend-engine**="": backend engine to run pipelines on (default: "auto-detect")

**--backend-http-proxy**="": if set, pass the environment variable down as "HTTP_PROXY" to steps

**--backend-https-proxy**="": if set, pass the environment variable down as "HTTPS_PROXY" to steps

**--backend-k8s-namespace**="": backend k8s namespace (default: "woodpecker")

**--backend-k8s-pod-annotations**="": backend k8s additional worker pod annotations

**--backend-k8s-pod-labels**="": backend k8s additional worker pod labels

**--backend-k8s-secctx-nonroot**: `run as non root` Kubernetes security context option

**--backend-k8s-storage-class**="": backend k8s storage class

**--backend-k8s-storage-rwx**: backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)

**--backend-k8s-volume-size**="": backend k8s volume size (default 10G) (default: "10G")

**--backend-local-temp-dir**="": set a different temp dir to clone workflows into (default: "/tmp")

**--backend-no-proxy**="": if set, pass the environment variable down as "NO_PROXY" to steps

**--commit-author-avatar**="": 

**--commit-author-email**="": 

**--commit-author-name**="": 

**--commit-branch**="": 

**--commit-message**="": 

**--commit-ref**="": 

**--commit-refspec**="": 

**--commit-sha**="": 

**--connect-retry-count**="": number of times to retry connecting to the server (default: 5)

**--connect-retry-delay**="": duration to wait before retrying to connect to the server (default: 2s)

**--env**="": 

**--forge-type**="": 

**--forge-url**="": 

**--local**: run from local directory

**--netrc-machine**="": 

**--netrc-password**="": 

**--netrc-username**="": 

**--network**="": external networks

**--pipeline-created**="":  (default: 0)

**--pipeline-event**="":  (default: "manual")

**--pipeline-finished**="":  (default: 0)

**--pipeline-number**="":  (default: 0)

**--pipeline-parent**="":  (default: 0)

**--pipeline-started**="":  (default: 0)

**--pipeline-status**="": 

**--pipeline-target**="": 

**--pipeline-url**="": 

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

**--step-name**="":  (default: 0)

**--system-name**="":  (default: "woodpecker")

**--system-platform**="": 

**--system-url**="":  (default: "https://github.com/woodpecker-ci/woodpecker")

**--timeout**="": pipeline timeout (default: 1h0m0s)

**--volumes**="": pipeline volumes

**--workflow-name**="":  (default: 0)

**--workflow-number**="":  (default: 0)

**--workspace-base**="":  (default: "/woodpecker")

**--workspace-path**="":  (default: "src")

## info

show information about the current user

## registry

manage registries

### add

adds a registry

**--hostname**="": registry hostname (default: "docker.io")

**--password**="": registry password

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--username**="": registry username

### rm

remove a registry

**--hostname**="": registry hostname (default: "docker.io")

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

### update

update a registry

**--hostname**="": registry hostname (default: "docker.io")

**--password**="": registry password

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--username**="": registry username

### info

display registry info

**--hostname**="": registry hostname (default: "docker.io")

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

### ls

list registries

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

## secret

manage secrets

### add

adds a secret

**--event**="": secret limited to these events

**--global**: global secret

**--image**="": secret limited to these images

**--name**="": secret name

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--value**="": secret value

### rm

remove a secret

**--global**: global secret

**--name**="": secret name

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

### update

update a secret

**--event**="": secret limited to these events

**--global**: global secret

**--image**="": secret limited to these images

**--name**="": secret name

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--value**="": secret value

### info

display secret info

**--global**: global secret

**--name**="": secret name

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

### ls

list secrets

**--global**: global secret

**--organization, --org**="": organization id or full-name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

## repo

manage repositories

### ls

list all repos

**--format**="": format output (default: "\x1b[33m{{ .FullName }}\x1b[0m (id: {{ .ID }}, forgeRemoteID: {{ .ForgeRemoteID }})")

**--org**="": filter by organization

### info

show repository details

**--format**="": format output (default: "Owner: {{ .Owner }}\nRepo: {{ .Name }}\nURL: {{ .ForgeURL }}\nConfig path: {{ .Config }}\nVisibility: {{ .Visibility }}\nPrivate: {{ .IsSCMPrivate }}\nTrusted: {{ .IsTrusted }}\nGated: {{ .IsGated }}\nClone url: {{ .Clone }}\nAllow pull-requests: {{ .AllowPullRequests }}\n")

### add

add a repository

### update

update a repository

**--config**="": repository configuration path (e.g. .woodpecker.yml)

**--gated**: repository is gated

**--pipeline-counter**="": repository starting pipeline number (default: 0)

**--timeout**="": repository timeout (default: 0s)

**--trusted**: repository is trusted

**--unsafe**: validate updating the pipeline-counter is unsafe

**--visibility**="": repository visibility

### rm

remove a repository

### repair

repair repository webhooks

### chown

assume ownership of a repository

### sync

synchronize the repository list

**--format**="": format output (default: "\x1b[33m{{ .FullName }}\x1b[0m (id: {{ .ID }}, forgeRemoteID: {{ .ForgeRemoteID }})")

## user

manage users

### ls

list all users

**--format**="": format output (default: "{{ .Login }}")

### info

show user details

**--format**="": format output (default: "User: {{ .Login }}\nEmail: {{ .Email }}")

### add

adds a user

### rm

remove a user

## lint

lint a pipeline configuration file

## log-level

get the logging level of the server, or set it with [level]

## cron

manage cron jobs

### add

add a cron job

**--branch**="": cron branch

**--name**="": cron name

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

### rm

remove a cron job

**--id**="": cron id

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

### update

update a cron job

**--branch**="": cron branch

**--id**="": cron id

**--name**="": cron name

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

### info

display info about a cron job

**--id**="": cron id

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

### ls

list cron jobs

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)
