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

## build

manage builds

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

show build history

**--branch**="": branch filter

**--event**="": event filter

**--format**="": format output (default: [33mBuild #{{ .Number }} [0m
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
)

**--limit**="": limit the list size (default: 25)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--status**="": status filter

**--token, -t**="": server auth token

### last

show latest build details

**--branch**="": branch name (default: master)

**--format**="": format output (default: Number: {{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### logs

show build logs

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show build details

**--format**="": format output (default: Number: {{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### stop

stop a build

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### start

start a build

**--log-level**="": set logging level (default: info)

**--param, -p**="": custom parameters to be injected into the job environment. Format: KEY=value

**--server, -s**="": server address

**--token, -t**="": server auth token

### approve

approve a build

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### decline

decline a build

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### queue

show build queue

**--format**="": format output (default: [33m{{ .FullName }} #{{ .Number }} [0m
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ps

show build steps

**--format**="": format output (default: [33mProc #{{ .PID }} [0m
Step: {{ .Name }}
State: {{ .State }}
)

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### create

create new build

**--branch**="": branch to create build from

**--format**="": format output (default: [33mBuild #{{ .Number }} [0m
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
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

**--format**="": format output (default: Number: {{ .Number }}
Status: {{ .Status }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}
Target: {{ .Deploy }}
)

**--log-level**="": set logging level (default: info)

**--param, -p**="": custom parameters to be injected into the job environment. Format: KEY=value

**--server, -s**="": server address

**--status**="": status filter (default: success)

**--token, -t**="": server auth token

## exec

execute a local build

**--backend-engine**="": backend engine to run pipelines on (default: auto-detect)

**--backend-k8s-namespace**="": backend k8s namespace (default: woodpecker)

**--backend-k8s-storage-class**="": backend k8s storage class

**--backend-k8s-storage-rwx**: backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)

**--backend-k8s-volume-size**="": backend k8s volume size (default 10G) (default: 10G)

**--build-created**="":  (default: 0)

**--build-event**="": 

**--build-finished**="":  (default: 0)

**--build-link**="": 

**--build-number**="":  (default: 0)

**--build-started**="":  (default: 0)

**--build-status**="": 

**--build-target**="": 

**--commit-author-avatar**="": 

**--commit-author-email**="": 

**--commit-author-name**="": 

**--commit-branch**="": 

**--commit-message**="": 

**--commit-ref**="": 

**--commit-refspec**="": 

**--commit-sha**="": 

**--env**="": 

**--job-number**="":  (default: 0)

**--local**: build from local directory

**--log-level**="": set logging level (default: info)

**--netrc-machine**="": 

**--netrc-password**="": 

**--netrc-username**="": 

**--network**="": external networks

**--parent-build-number**="":  (default: 0)

**--prev-build-created**="":  (default: 0)

**--prev-build-event**="": 

**--prev-build-finished**="":  (default: 0)

**--prev-build-link**="": 

**--prev-build-number**="":  (default: 0)

**--prev-build-started**="":  (default: 0)

**--prev-build-status**="": 

**--prev-commit-author-avatar**="": 

**--prev-commit-author-email**="": 

**--prev-commit-author-name**="": 

**--prev-commit-branch**="": 

**--prev-commit-message**="": 

**--prev-commit-ref**="": 

**--prev-commit-refspec**="": 

**--prev-commit-sha**="": 

**--privileged**="": privileged plugins (default: [plugins/docker plugins/gcr plugins/ecr woodpeckerci/plugin-docker woodpeckerci/plugin-docker-buildx])

**--repo-link**="": 

**--repo-name**="": 

**--repo-private**="": 

**--repo-remote-url**="": 

**--server, -s**="": server address

**--system-link**="":  (default: https://github.com/cncd/pipec)

**--system-name**="":  (default: pipec)

**--system-platform**="": 

**--timeout**="": build timeout (default: 1h0m0s)

**--token, -t**="": server auth token

**--volumes**="": build volumes

**--workspace-base**="":  (default: /woodpecker)

**--workspace-path**="":  (default: src)

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

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--username**="": registry username

### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--log-level**="": set logging level (default: info)

**--password**="": registry password

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--username**="": registry username

### info

display registry info

**--hostname**="": registry hostname (default: docker.io)

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list registries

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

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

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--value**="": secret value

### rm

remove a secret

**--global**: global secret

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--organization**="": organization name (e.g. octocat)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

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

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

**--value**="": secret value

### info

display secret info

**--global**: global secret

**--log-level**="": set logging level (default: info)

**--name**="": secret name

**--organization**="": organization name (e.g. octocat)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list secrets

**--global**: global secret

**--log-level**="": set logging level (default: info)

**--organization**="": organization name (e.g. octocat)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

## repo

manage repositories

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list all repos

**--format**="": format output (default: {{ .FullName }})

**--log-level**="": set logging level (default: info)

**--org**="": filter by organization

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show repository details

**--format**="": format output (default: Owner: {{ .Owner }}
Repo: {{ .Name }}
Type: {{ .SCMKind }}
Config: {{ .Config }}
Visibility: {{ .Visibility }}
Private: {{ .IsSCMPrivate }}
Trusted: {{ .IsTrusted }}
Gated: {{ .IsGated }}
Remote: {{ .Clone }}
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

**--build-counter**="": repository starting build number (default: 0)

**--config**="": repository configuration path (e.g. .woodpecker.yml)

**--gated**: repository is gated

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--timeout**="": repository timeout (default: 0s)

**--token, -t**="": server auth token

**--trusted**: repository is trusted

**--unsafe**: validate updating the build-counter is unsafe

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

**--format**="": format output (default: {{ .FullName }})

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

**--format**="": format output (default: {{ .Login }})

**--log-level**="": set logging level (default: info)

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

show user details

**--format**="": format output (default: User: {{ .Login }}
Email: {{ .Email }})

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

adds a cron

**--branch**="": cron branch

**--log-level**="": set logging level (default: info)

**--name**="": cron name

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--schedule**="": cron schedule

**--server, -s**="": server address

**--token, -t**="": server auth token

### rm

remove a cron

**--id**="": cron id

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### update

update a cron

**--branch**="": cron branch

**--id**="": cron id

**--log-level**="": set logging level (default: info)

**--name**="": cron name

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--schedule**="": cron schedule

**--server, -s**="": server address

**--token, -t**="": server auth token

### info

display cron info

**--id**="": cron id

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token

### ls

list registries

**--log-level**="": set logging level (default: info)

**--repository, --repo**="": repository name (e.g. octocat/hello-world)

**--server, -s**="": server address

**--token, -t**="": server auth token
