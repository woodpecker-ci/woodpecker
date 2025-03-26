# CLI

# NAME

woodpecker-cli - command line utility

# SYNOPSIS

woodpecker-cli

```
[--config|-c]=[value]
[--disable-update-check]
[--log-file]=[value]
[--log-level]=[value]
[--nocolor]
[--pretty]
[--server|-s]=[value]
[--token|-t]=[value]
```

# DESCRIPTION

Woodpecker command line utility

**Usage**:

```
woodpecker-cli [GLOBAL OPTIONS] [command [COMMAND OPTIONS]] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--config, -c**="": path to config file

**--disable-update-check**: disable update check

**--log-file**="": Output destination for logs. 'stdout' and 'stderr' can be used as special keywords. (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too

**--pretty**: enable pretty-printed debug output

**--server, -s**="": server address

**--token, -t**="": server auth token

# COMMANDS

## admin

administer server settings

### registry

manage global registries

#### add

adds a registry

**--hostname**="": registry hostname (default: docker.io)

**--password**="": registry password

**--username**="": registry username

#### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

#### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--password**="": registry password

**--username**="": registry username

#### info

display registry info

**--hostname**="": registry hostname (default: docker.io)

#### ls

list registries

## org

manage organizations

### registry

manage organization registries

#### add

adds a registry

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--password**="": registry password

**--username**="": registry username

#### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

#### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--password**="": registry password

**--username**="": registry username

#### info

display registry info

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

#### ls

list registries

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

## repo

manage repositories

### ls

list all repos

**--format**="": format output (default: [33m{{ .FullName }}[0m (id: {{ .ID }}, forgeRemoteID: {{ .ForgeRemoteID }}))

**--org**="": filter by organization

### info

show repository details

**--format**="": format output (default: Owner: {{ .Owner }}
Repo: {{ .Name }}
URL: {{ .ForgeURL }}
Config path: {{ .Config }}
Visibility: {{ .Visibility }}
Private: {{ .IsSCMPrivate }}
Trusted: {{ .IsTrusted }}
Gated: {{ .IsGated }}
Require approval for: {{ .RequireApproval }}
Clone url: {{ .Clone }}
Allow pull-requests: {{ .AllowPullRequests }}
)

### add

add a repository

### update

update a repository

**--config**="": repository configuration path (e.g. .woodpecker.yml)

**--gated**: [deprecated] repository is gated

**--pipeline-counter**="": repository starting pipeline number (default: 0)

**--require-approval**="": repository requires approval for

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

**--format**="": format output (default: [33m{{ .FullName }}[0m (id: {{ .ID }}, forgeRemoteID: {{ .ForgeRemoteID }}))

### registry

manage registries

#### add

adds a registry

**--hostname**="": registry hostname (default: docker.io)

**--password**="": registry password

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--username**="": registry username

#### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--password**="": registry password

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--username**="": registry username

#### info

display registry info

**--hostname**="": registry hostname (default: docker.io)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### ls

list registries

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

## pipeline

manage pipelines

### ls

show pipeline history

**--branch**="": branch filter

**--event**="": event filter

**--limit**="": limit the list size (default: 25)

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers

**--status**="": status filter

### last

show latest pipeline details

**--branch**="": branch name (default: main)

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers

### logs

show pipeline logs

### info

show pipeline details

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers

### stop

stop a pipeline

### start

start a pipeline

**--param, -p**="": custom parameters to be injected into the step environment. Format: KEY=value (default: [])

### approve

approve a pipeline

### decline

decline a pipeline

### queue

show pipeline queue

**--format**="": format output (default: [33m{{ .FullName }} #{{ .Number }} [0m
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
)

### ps

show pipeline steps

**--format**="": format output (default: [33m{{ .workflow.Name }} > {{ .step.Name }} (#{{ .step.PID }}):[0m
Step: {{ .step.Name }}
Started: {{ .step.Started }}
Stopped: {{ .step.Stopped }}
Type: {{ .step.Type }}
State: {{ .step.State }}
)

### create

create new pipeline

**--branch**="": branch to create pipeline from

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers

**--var**="": key=value (default: [])

## log

manage logs

### purge

purge a log

## deploy

trigger a pipeline with the 'deployment' event

**--branch**="": branch filter

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

**--param, -p**="": custom parameters to be injected into the step environment. Format: KEY=value (default: [])

**--status**="": status filter (default: success)

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

**--backend-k8s-allow-native-secrets**: whether to allow existing Kubernetes secrets to be referenced from steps

**--backend-k8s-namespace**="": backend k8s namespace (default: woodpecker)

**--backend-k8s-pod-annotations**="": backend k8s additional Agent-wide worker pod annotations

**--backend-k8s-pod-annotations-allow-from-step**: whether to allow using annotations from step's backend options

**--backend-k8s-pod-image-pull-secret-names**="": backend k8s pull secret names for private registries (default: [regcred])

**--backend-k8s-pod-labels**="": backend k8s additional Agent-wide worker pod labels

**--backend-k8s-pod-labels-allow-from-step**: whether to allow using labels from step's backend options

**--backend-k8s-pod-node-selector**="": backend k8s Agent-wide worker pod node selector

**--backend-k8s-secctx-nonroot**: `run as non root` Kubernetes security context option

**--backend-k8s-storage-class**="": backend k8s storage class

**--backend-k8s-storage-rwx**: backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)

**--backend-k8s-volume-size**="": backend k8s volume size (default 10G) (default: 10G)

**--backend-local-temp-dir**="": set a different temp dir to clone workflows into (default: /tmp/nix-shell.OgDG7Z)

**--backend-no-proxy**="": if set, pass the environment variable down as "NO_PROXY" to steps

**--commit-author-avatar**="":

**--commit-author-email**="":

**--commit-author-name**="":

**--commit-branch**="":

**--commit-message**="":

**--commit-ref**="":

**--commit-refspec**="":

**--commit-sha**="":

**--env**="": (default: [])

**--forge-type**="":

**--forge-url**="":

**--local**: run from local directory

**--netrc-machine**="":

**--netrc-password**="":

**--netrc-username**="":

**--network**="": external networks (default: [])

**--pipeline-created**="": (default: 0)

**--pipeline-deploy-task**="":

**--pipeline-deploy-to**="":

**--pipeline-event**="": (default: manual)

**--pipeline-finished**="": (default: 0)

**--pipeline-number**="": (default: 0)

**--pipeline-parent**="": (default: 0)

**--pipeline-started**="": (default: 0)

**--pipeline-status**="":

**--pipeline-url**="":

**--prev-commit-author-avatar**="":

**--prev-commit-author-email**="":

**--prev-commit-author-name**="":

**--prev-commit-branch**="":

**--prev-commit-message**="":

**--prev-commit-ref**="":

**--prev-commit-refspec**="":

**--prev-commit-sha**="":

**--prev-pipeline-created**="": (default: 0)

**--prev-pipeline-event**="":

**--prev-pipeline-finished**="": (default: 0)

**--prev-pipeline-number**="": (default: 0)

**--prev-pipeline-started**="": (default: 0)

**--prev-pipeline-status**="":

**--prev-pipeline-url**="":

**--privileged**="": privileged plugins (default: [plugins/docker plugins/gcr plugins/ecr woodpeckerci/plugin-docker-buildx codeberg.org/woodpecker-plugins/docker-buildx])

**--repo**="": full repo name

**--repo-clone-ssh-url**="":

**--repo-clone-url**="":

**--repo-path**="": path to local repository

**--repo-private**="":

**--repo-remote-id**="":

**--repo-trusted**:

**--repo-url**="":

**--step-name**="": (default: 0)

**--system-name**="": (default: woodpecker)

**--system-platform**="":

**--system-url**="": (default: https://github.com/woodpecker-ci/woodpecker)

**--timeout**="": pipeline timeout (default: 1h0m0s)

**--volumes**="": pipeline volumes (default: [])

**--workflow-name**="": (default: 0)

**--workflow-number**="": (default: 0)

**--workspace-base**="": (default: /woodpecker)

**--workspace-path**="": (default: src)

## info

show information about the current user

## registry

manage registries

### add

adds a registry

**--hostname**="": registry hostname (default: docker.io)

**--password**="": registry password

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--username**="": registry username

### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--password**="": registry password

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--username**="": registry username

### info

display registry info

**--hostname**="": registry hostname (default: docker.io)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

### ls

list registries

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

## secret

manage secrets

### add

adds a secret

**--event**="": secret limited to these events (default: [])

**--global**: global secret

**--image**="": secret limited to these images (default: [])

**--name**="": secret name

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--value**="": secret value

### rm

remove a secret

**--global**: global secret

**--name**="": secret name

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

### update

update a secret

**--event**="": secret limited to these events (default: [])

**--global**: global secret

**--image**="": secret limited to these images (default: [])

**--name**="": secret name

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--value**="": secret value

### info

display secret info

**--global**: global secret

**--name**="": secret name

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

### ls

list secrets

**--global**: global secret

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

## user

manage users

### ls

list all users

**--format**="": format output (default: {{ .Login }})

### info

show user details

**--format**="": format output (default: User: {{ .Login }}
Email: {{ .Email }})

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

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

### rm

remove a cron job

**--id**="": cron id

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

### update

update a cron job

**--branch**="": cron branch

**--id**="": cron id

**--name**="": cron name

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

### info

display info about a cron job

**--id**="": cron id

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

### ls

list cron jobs

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

## setup

setup the woodpecker-cli for the first time

**--server**="": The URL of the woodpecker server

**--token**="": The token to authenticate with the woodpecker server

## update

update the woodpecker-cli to the latest version

**--force**: force update even if the latest version is already installed
