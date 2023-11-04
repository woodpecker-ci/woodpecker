# CLI

# NAME

woodpecker-cli - command line utility

# SYNOPSIS

woodpecker-cli

```
[--log-level]=[value]
[--server|-s]=[value]
[--token|-t]=[value]
```

**Usage**:

```
woodpecker-cli [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

# COMMANDS

## pipeline, build

manage pipelines

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

show pipeline history

**--branch**="": branch filter

**--event**="": event filter

**--format**="": format output (default: `Pipeline #{{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
)`

**--limit**="": limit the list size (default: 0)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--status**="": status filter

**--token, -t**="": server auth token

### last

show latest pipeline details

**--branch**="": branch name (default: master)

**--format**="": format output (default: Number: `{{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}`
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### logs

show pipeline logs

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show pipeline details

**--format**="": format output (default: Number: `{{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}`
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### stop

stop a pipeline

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### start

start a pipeline

**--log-level**="": set logging level (default: info)

**--param, -p**="": custom parameters to be injected into the step environment. Format: KEY=value

**--server, -s**="": server address

**--token, -t**="": server auth token

### approve

approve a pipeline

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### decline

decline a pipeline

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### queue

show pipeline queue

**--format**="": format output (default: `{{ .FullName }} #{{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}`
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ps

show pipeline steps

**--format**="": format output (default: `Step #{{ .PID }} Step: {{ .Name }}
State: {{ .State }}`
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### create

create new pipeline

**--branch**="": branch to create pipeline from

**--format**="": format output (default: `Pipeline #{{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}`
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--var**="": key=value

## log

manage logs

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### purge

purge a log

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

## deploy

deploy code

**--branch**="": branch filter (default: master)

**--event**="": event filter (default: push)

**--format**="": format output (default: Number: `{{ .Number }}
Status: {{ .Status }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}
Target: {{ .Deploy }}`
)

**--log-level**="": set logging level (default: info)

**--param, -p**="": custom parameters to be injected into the step environment. Format: KEY=value

**--server, -s**="": server address

**--status**="": status filter (default: success)

**--token, -t**="": server auth token

## exec

execute a local pipeline

**--backend-docker-ipv6**: backend docker enable IPV6

**--backend-docker-network**="": backend docker network

**--backend-docker-volumes**="": backend docker volumes (comma separated)

**--backend-engine**="": backend engine to run pipelines on (default: auto-detect)

**--backend-k8s-namespace**="": backend k8s namespace (default: woodpecker)

**--backend-k8s-pod-annotations**="": backend k8s additional worker pod annotations

**--backend-k8s-pod-labels**="": backend k8s additional worker pod labels

**--backend-k8s-storage-class**="": backend k8s storage class

**--backend-k8s-storage-rwx**: backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)

**--backend-k8s-volume-size**="": backend k8s volume size (default 10G) (default: 10G)

**--backend-ssh-address**="": backend ssh address

**--backend-ssh-key**="": backend ssh key file

**--backend-ssh-key-password**="": backend ssh key password

**--backend-ssh-password**="": backend ssh password

**--backend-ssh-user**="": backend ssh user

**--commit-author-avatar**="":

**--commit-author-email**="":

**--commit-author-name**="":

**--commit-branch**="":

**--commit-message**="":

**--commit-ref**="":

**--commit-refspec**="":

**--commit-sha**="":

**--env**="":

**--forge-type**="":

**--forge-url**="":

**--local**: run from local directory

**--log-level**="": set logging level (default: info)

**--netrc-machine**="":

**--netrc-password**="":

**--netrc-username**="":

**--network**="": external networks

**--pipeline-created**="": (default: 0)

**--pipeline-event**="": (default: manual)

**--pipeline-finished**="": (default: 0)

**--pipeline-link**="":

**--pipeline-number**="": (default: 0)

**--pipeline-parent**="": (default: 0)

**--pipeline-started**="": (default: 0)

**--pipeline-status**="":

**--pipeline-target**="":

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

**--prev-pipeline-link**="":

**--prev-pipeline-number**="": (default: 0)

**--prev-pipeline-started**="": (default: 0)

**--prev-pipeline-status**="":

**--privileged**="": privileged plugins (default: "plugins/docker", "plugins/gcr", "plugins/ecr", "woodpeckerci/plugin-docker-buildx")

**--repo**="": full repo name

**--repo-clone-url**="":

**--repo-link**="":

**--repo-private**="":

**--repo-remote-id**="":

**--repo-trusted**:

**--server, -s**="": server address

**--step-name**="": (default: 0)

**--system-link**="": (default: https://github.com/woodpecker-ci/woodpecker)

**--system-name**="": (default: woodpecker)

**--system-platform**="":

**--timeout**="": pipeline timeout (default: 0s)

**--token, -t**="": server auth token

**--volumes**="": pipeline volumes

**--workflow-name**="": (default: 0)

**--workflow-number**="": (default: 0)

**--workspace-base**="": (default: /woodpecker)

**--workspace-path**="": (default: src)

## info

show information about the current user

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

## registry

manage registries

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

adds a registry

**--hostname**="": registry hostname (default: docker.io)

**--log-level**="": set logging level (default: info)

**--password**="": registry password

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--username**="": registry username

### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--log-level**="": set logging level (default: info)

**--password**="": registry password

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--username**="": registry username

### info

display registry info

**--hostname**="": registry hostname (default: docker.io)

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list registries

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

## secret

manage secrets

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

adds a secret

**--event**="": secret limited to these events

**--global**: global secret

**--image**="": secret limited to these images

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--organization**="": organization name (e.g. octocat)

**--plugins-only**: secret limited to plugins

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--value**="": secret value

### rm

remove a secret

**--global**: global secret

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--organization**="": organization name (e.g. octocat)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a secret

**--event**="": secret limited to these events

**--global**: global secret

**--image**="": secret limited to these images

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--organization**="": organization name (e.g. octocat)

**--plugins-only**: secret limited to plugins

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--value**="": secret value

### info

display secret info

**--global**: global secret

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--organization**="": organization name (e.g. octocat)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list secrets

**--global**: global secret

**--log-level**="": set logging level (default: info)

**--organization**="": organization name (e.g. octocat)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

## repo

manage repositories

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list all repos

**--format**="": format output (default: `{{ .FullName }} (id: {{ .ID }})`)

**--log-level**="": set logging level (default: info)

**--org**="": filter by organization

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show repository details

**--format**="": format output (default: Owner: `{{ .Owner }}
Repo: {{ .Name }}
Link: {{ .Link }}
Config path: {{ .Config }}
Visibility: {{ .Visibility }}
Private: {{ .IsSCMPrivate }}
Trusted: {{ .IsTrusted }}
Gated: {{ .IsGated }}
Clone url: {{ .Clone }}
Allow pull-requests: {{ .AllowPullRequests }}`
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

add a repository

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a repository

**--config**="": repository configuration path (e.g. .woodpecker.yml)

**--gated**: repository is gated

**--log-level**="": set logging level (default: info)

**--pipeline-counter**="": repository starting pipeline number (default: 0)

**--server, -s**="": server address

**--timeout**="": repository timeout (default: 0s)

**--token, -t**="": server auth token

**--trusted**: repository is trusted

**--unsafe**: validate updating the pipeline-counter is unsafe

**--visibility**="": repository visibility

### rm

remove a repository

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### repair

repair repository webhooks

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### chown

assume ownership of a repository

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### sync

synchronize the repository list

**--format**="": format output (default: `{{ .FullName }} (id: {{ .ID }})`)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

## user

manage users

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list all users

**--format**="": format output (default: `{{ .Login }}`)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show user details

**--format**="": format output (default: User: `{{ .Login }}
Email: {{ .Email }}`)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

adds a user

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### rm

remove a user

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

## lint

lint a pipeline configuration file

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

## log-level

get the logging level of the server, or set it with [level]

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

## cron

manage cron jobs

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### add

add a cron job

**--branch**="": cron branch

**--log-level**="": set logging level (default: info)

**--name**="": cron name

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

**--server, -s**="": server address

**--token, -t**="": server auth token

### rm

remove a cron job

**--id**="": cron id

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a cron job

**--branch**="": cron branch

**--id**="": cron id

**--log-level**="": set logging level (default: info)

**--name**="": cron name

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

display info about a cron job

**--id**="": cron id

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list cron jobs

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository id or full-name (e.g. 134 or octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token
