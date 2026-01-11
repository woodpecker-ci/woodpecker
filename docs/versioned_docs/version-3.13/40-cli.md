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
[--skip-verify]
[--socks-proxy-off]
[--socks-proxy]=[value]
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

**--disable-update-check**: disable update check (default: false)

**--log-file**="": Output destination for logs. 'stdout' and 'stderr' can be used as special keywords. (default: stderr)

**--log-level**="": set logging level (default: info)

**--nocolor**: disable colored debug output, only has effect if pretty output is set too (default: false)

**--pretty**: enable pretty-printed debug output (default: true)

**--server, -s**="": server address

**--skip-verify**: skip ssl verification (default: false)

**--socks-proxy**="": socks proxy address

**--socks-proxy-off**: socks proxy ignored (default: false)

**--token, -t**="": server auth token


# COMMANDS

## admin

manage server settings

### log-level

retrieve log level from server, or set it with [level]

### org

manage organizations

#### ls

list organizations

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nOrganization ID: {{ .ID }}\n)

### registry

manage global registries

#### add

add a registry

**--hostname**="": registry hostname (default: docker.io)

**--password**="": registry password

**--username**="": registry username

#### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

#### ls

list registries

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Address }} \x1b[0m\nUsername: {{ .Username }}\nEmail: {{ .Email }}\n)

#### show

show registry information

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Address }} \x1b[0m\nUsername: {{ .Username }}\nEmail: {{ .Email }}\n)

**--hostname**="": registry hostname (default: docker.io)

#### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--password**="": registry password

**--username**="": registry username

### secret

manage global secrets

#### add

add a secret

**--event**="": secret limited to these events

**--image**="": secret limited to these images

**--name**="": secret name

**--value**="": secret value

#### rm

remove a secret

**--name**="": secret name

#### ls

list secrets

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nEvents: {{ list .Events }}\n{{- if .Images }}\nImages: {{ list .Images }}\n{{- else }}\nImages: <any>\n{{- end }}\n)

#### show

show secret information

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nEvents: {{ list .Events }}\n{{- if .Images }}\nImages: {{ list .Images }}\n{{- else }}\nImages: <any>\n{{- end }}\n)

**--name**="": secret name

#### update

update a secret

**--event**="": secret limited to these events

**--image**="": secret limited to these images

**--name**="": secret name

**--value**="": secret value

### user

manage users

#### add

add a user

#### ls

list all users

**--format**="": format output (default: {{ .Login }})

#### rm

remove a user

#### show

show user information

**--format**="": format output (default: User: {{ .Login }}\nEmail: {{ .Email }})

## exec

execute a local pipeline

**--backend-docker-api-version**="": the version of the API to reach, leave empty for latest.

**--backend-docker-cert**="": path to load the TLS certificates for connecting to docker server

**--backend-docker-host**="": path to docker socket or url to the docker server

**--backend-docker-ipv6**: backend docker enable IPV6 (default: false)

**--backend-docker-limit-cpu-quota**="": impose a cpu quota (default: 0)

**--backend-docker-limit-cpu-set**="": set the cpus allowed to execute containers

**--backend-docker-limit-cpu-shares**="": change the cpu shares (default: 0)

**--backend-docker-limit-mem**="": maximum memory allowed in bytes (default: 0)

**--backend-docker-limit-mem-swap**="": maximum memory used for swap in bytes (default: 0)

**--backend-docker-limit-shm-size**="": docker /dev/shm allowed in bytes (default: 0)

**--backend-docker-network**="": backend docker network

**--backend-docker-tls-verify**: enable or disable TLS verification for connecting to docker server (default: true)

**--backend-docker-volumes**="": backend docker volumes (comma separated)

**--backend-engine**="": backend engine to run pipelines on (default: auto-detect)

**--backend-http-proxy**="": if set, pass the environment variable down as "HTTP_PROXY" to steps

**--backend-https-proxy**="": if set, pass the environment variable down as "HTTPS_PROXY" to steps

**--backend-k8s-allow-native-secrets**: whether to allow existing Kubernetes secrets to be referenced from steps (default: false)

**--backend-k8s-namespace**="": backend k8s namespace, if used with WOODPECKER_BACKEND_K8S_NAMESPACE_PER_ORGANIZATION, this will be the prefix for the namespace appended with the organization name. (default: woodpecker)

**--backend-k8s-namespace-per-org**: Whether to enable namespace segregation per organization feature. When enabled, Woodpecker will create the Kubernetes resources to separated Kubernetes namespaces per Woodpecker organization. (default: false)

**--backend-k8s-pod-affinity**="": backend k8s Agent-wide worker pod affinity, in YAML format

**--backend-k8s-pod-affinity-allow-from-step**: whether to allow using affinity from step's backend options (default: false)

**--backend-k8s-pod-annotations**="": backend k8s additional Agent-wide worker pod annotations

**--backend-k8s-pod-annotations-allow-from-step**: whether to allow using annotations from step's backend options (default: false)

**--backend-k8s-pod-image-pull-secret-names**="": backend k8s pull secret names for private registries

**--backend-k8s-pod-labels**="": backend k8s additional Agent-wide worker pod labels

**--backend-k8s-pod-labels-allow-from-step**: whether to allow using labels from step's backend options (default: false)

**--backend-k8s-pod-node-selector**="": backend k8s Agent-wide worker pod node selector

**--backend-k8s-pod-tolerations**="": backend k8s Agent-wide worker pod tolerations

**--backend-k8s-pod-tolerations-allow-from-step**: whether to allow using tolerations from step's backend options (default: true)

**--backend-k8s-priority-class**="": which kubernetes priority class to assign to created job pods

**--backend-k8s-secctx-nonroot**: `run as non root` Kubernetes security context option (default: false)

**--backend-k8s-storage-class**="": backend k8s storage class

**--backend-k8s-storage-rwx**: backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true) (default: true)

**--backend-k8s-volume-size**="": backend k8s volume size (default 10G) (default: 10G)

**--backend-local-temp-dir**="": set a different temp dir to clone workflows into (default: system temporary directory)

**--backend-no-proxy**="": if set, pass the environment variable down as "NO_PROXY" to steps

**--commit-author-avatar**="": Set the metadata environment variable "CI_COMMIT_AUTHOR_AVATAR".

**--commit-author-email**="": Set the metadata environment variable "CI_COMMIT_AUTHOR_EMAIL".

**--commit-author-name**="": Set the metadata environment variable "CI_COMMIT_AUTHOR".

**--commit-branch**="": Set the metadata environment variable "CI_COMMIT_BRANCH". (default: main)

**--commit-message**="": Set the metadata environment variable "CI_COMMIT_MESSAGE".

**--commit-pull-labels**="": Set the metadata environment variable "CI_COMMIT_PULL_REQUEST_LABELS".

**--commit-pull-milestone**="": Set the metadata environment variable "CI_COMMIT_PULL_REQUEST_MILESTONE".

**--commit-ref**="": Set the metadata environment variable "CI_COMMIT_REF".

**--commit-refspec**="": Set the metadata environment variable "CI_COMMIT_REFSPEC".

**--commit-release-is-pre**: Set the metadata environment variable "CI_COMMIT_PRERELEASE". (default: false)

**--commit-sha**="": Set the metadata environment variable "CI_COMMIT_SHA".

**--env**="": Set the metadata environment variable "CI_ENV".

**--forge-type**="": Set the metadata environment variable "CI_FORGE_TYPE".

**--forge-url**="": Set the metadata environment variable "CI_FORGE_URL".

**--local**: run from local directory (default: true)

**--metadata-file**="": path to pipeline metadata file (normally downloaded from UI). Parameters can be adjusted by applying additional cli flags

**--netrc-machine**="": 

**--netrc-password**="": 

**--netrc-username**="": 

**--network**="": external networks

**--pipeline-changed-files**="": Set the metadata environment variable "CI_PIPELINE_FILES", either json formatted list of strings, or comma separated string list.

**--pipeline-created**="": Set the metadata environment variable "CI_PIPELINE_CREATED". (default: 0)

**--pipeline-deploy-task**="": Set the metadata environment variable "CI_PIPELINE_DEPLOY_TASK".

**--pipeline-deploy-to**="": Set the metadata environment variable "CI_PIPELINE_DEPLOY_TARGET".

**--pipeline-event**="": Set the metadata environment variable "CI_PIPELINE_EVENT". (default: manual)

**--pipeline-number**="": Set the metadata environment variable "CI_PIPELINE_NUMBER". (default: 0)

**--pipeline-parent**="": Set the metadata environment variable "CI_PIPELINE_PARENT". (default: 0)

**--pipeline-started**="": Set the metadata environment variable "CI_PIPELINE_STARTED". (default: 0)

**--pipeline-url**="": Set the metadata environment variable "CI_PIPELINE_FORGE_URL".

**--plugins-privileged**="": Allow plugins to run in privileged mode, if environment variable is defined but empty there will be none

**--prev-commit-author-avatar**="": Set the metadata environment variable "CI_PREV_COMMIT_AUTHOR_AVATAR".

**--prev-commit-author-email**="": Set the metadata environment variable "CI_PREV_COMMIT_AUTHOR_EMAIL".

**--prev-commit-author-name**="": Set the metadata environment variable "CI_PREV_COMMIT_AUTHOR".

**--prev-commit-branch**="": Set the metadata environment variable "CI_PREV_COMMIT_BRANCH".

**--prev-commit-message**="": Set the metadata environment variable "CI_PREV_COMMIT_MESSAGE".

**--prev-commit-ref**="": Set the metadata environment variable "CI_PREV_COMMIT_REF".

**--prev-commit-refspec**="": Set the metadata environment variable "CI_PREV_COMMIT_REFSPEC".

**--prev-commit-sha**="": Set the metadata environment variable "CI_PREV_COMMIT_SHA".

**--prev-pipeline-created**="": Set the metadata environment variable "CI_PREV_PIPELINE_CREATED". (default: 0)

**--prev-pipeline-deploy-task**="": Set the metadata environment variable "CI_PREV_PIPELINE_DEPLOY_TASK".

**--prev-pipeline-deploy-to**="": Set the metadata environment variable "CI_PREV_PIPELINE_DEPLOY_TARGET".

**--prev-pipeline-event**="": Set the metadata environment variable "CI_PREV_PIPELINE_EVENT".

**--prev-pipeline-finished**="": Set the metadata environment variable "CI_PREV_PIPELINE_FINISHED". (default: 0)

**--prev-pipeline-number**="": Set the metadata environment variable "CI_PREV_PIPELINE_NUMBER". (default: 0)

**--prev-pipeline-started**="": Set the metadata environment variable "CI_PREV_PIPELINE_STARTED". (default: 0)

**--prev-pipeline-status**="": Set the metadata environment variable "CI_PREV_PIPELINE_STATUS".

**--prev-pipeline-url**="": Set the metadata environment variable "CI_PREV_PIPELINE_FORGE_URL".

**--repo**="": Set the full name to derive metadata environment variables "CI_REPO", "CI_REPO_NAME" and "CI_REPO_OWNER".

**--repo-clone-ssh-url**="": Set the metadata environment variable "CI_REPO_CLONE_SSH_URL".

**--repo-clone-url**="": Set the metadata environment variable "CI_REPO_CLONE_URL".

**--repo-default-branch**="": Set the metadata environment variable "CI_REPO_DEFAULT_BRANCH". (default: main)

**--repo-path**="": path to local repository

**--repo-private**="": Set the metadata environment variable "CI_REPO_PRIVATE".

**--repo-remote-id**="": Set the metadata environment variable "CI_REPO_REMOTE_ID".

**--repo-trusted-network**: Set the metadata environment variable "CI_REPO_TRUSTED_NETWORK". (default: false)

**--repo-trusted-security**: Set the metadata environment variable "CI_REPO_TRUSTED_SECURITY". (default: false)

**--repo-trusted-volumes**: Set the metadata environment variable "CI_REPO_TRUSTED_VOLUMES". (default: false)

**--repo-url**="": Set the metadata environment variable "CI_REPO_URL".

**--secrets**="": map of secrets, ex. 'secret="val",secret2="value2"'

**--secrets-file**="": path to yaml file with secrets map

**--system-host**="": Set the metadata environment variable "CI_SYSTEM_HOST".

**--system-name**="": Set the metadata environment variable "CI_SYSTEM_NAME". (default: woodpecker)

**--system-platform**="": Set the metadata environment variable "CI_SYSTEM_PLATFORM".

**--system-url**="": Set the metadata environment variable "CI_SYSTEM_URL". (default: https://github.com/woodpecker-ci/woodpecker)

**--timeout**="": pipeline timeout (default: 1h0m0s)

**--volumes**="": pipeline volumes

**--workflow-name**="": Set the metadata environment variable "CI_WORKFLOW_NAME".

**--workflow-number**="": Set the metadata environment variable "CI_WORKFLOW_NUMBER". (default: 0)

**--workspace-base**="":  (default: /woodpecker)

**--workspace-path**="":  (default: src)

## info

show information about the current user

**--format**="": format output (deprecated) (default: User: {{ .Login }}\nEmail: {{ .Email }})

## lint

lint a pipeline configuration file

**--plugins-privileged**="": allow plugins to run in privileged mode, if set empty, there is no

**--plugins-trusted-clone**="": plugins that are trusted to handle Git credentials in cloning steps (default: "docker.io/woodpeckerci/plugin-git:2.7.0", "docker.io/woodpeckerci/plugin-git", "quay.io/woodpeckerci/plugin-git")

**--strict**: treat warnings as errors (default: false)

## org

manage organizations

### registry

manage organization registries

#### add

add a registry

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--password**="": registry password

**--username**="": registry username

#### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

#### ls

list registries

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Address }} \x1b[0m\nUsername: {{ .Username }}\nEmail: {{ .Email }}\n)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

#### show

show registry information

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Address }} \x1b[0m\nUsername: {{ .Username }}\nEmail: {{ .Email }}\n)

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

#### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--password**="": registry password

**--username**="": registry username

### secret

manage secrets

#### add

add a secret

**--event**="": secret limited to these events

**--image**="": secret limited to these images

**--name**="": secret name

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--value**="": secret value

#### rm

remove a secret

**--name**="": secret name

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

#### ls

list secrets

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nEvents: {{ list .Events }}\n{{- if .Images }}\nImages: {{ list .Images }}\n{{- else }}\nImages: <any>\n{{- end }}\n)

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

#### show

show secret information

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nEvents: {{ list .Events }}\n{{- if .Images }}\nImages: {{ list .Images }}\n{{- else }}\nImages: <any>\n{{- end }}\n)

**--name**="": secret name

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

#### update

update a secret

**--event**="": limit secret to these event

**--image**="": limit secret to these image

**--name**="": secret name

**--organization, --org**="": organization id or full name (e.g. 123 or octocat)

**--value**="": secret value

## pipeline

manage pipelines

### approve

approve a pipeline

### create

create new pipeline

**--branch**="": branch to create pipeline from

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers (default: false)

**--var**="": key=value

### decline

decline a pipeline

### deploy

trigger a pipeline with the 'deployment' event

**--branch**="": branch filter

**--event**="": event filter (default: push)

**--format**="": format output (default: Number: {{ .Number }}\nStatus: {{ .Status }}\nCommit: {{ .Commit }}\nBranch: {{ .Branch }}\nRef: {{ .Ref }}\nMessage: {{ .Message }}\nAuthor: {{ .Author }}\nTarget: {{ .Deploy }}\n)

**--param, -p**="": custom parameters to inject into the step environment. Format: KEY=value

**--status**="": status filter (default: success)

### last

show latest pipeline information

**--branch**="": branch name (default: main)

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers (default: false)

### ls

show pipeline history

**--after**="": only return pipelines after this date (RFC3339)

**--before**="": only return pipelines before this date (RFC3339)

**--branch**="": branch filter

**--event**="": event filter

**--limit**="": limit the list size (default: 25)

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers (default: false)

**--status**="": status filter

### log

manage logs

#### purge

purge a log

#### show

show pipeline logs

### ps

show pipeline steps

**--format**="": format output (default: \x1b[33m{{ .workflow.Name }} > {{ .step.Name }} (#{{ .step.PID }}):\x1b[0m\nStep: {{ .step.Name }}\nStarted: {{ .step.Started }}\nStopped: {{ .step.Stopped }}\nType: {{ .step.Type }}\nState: {{ .step.State }}\n)

### purge

purge pipelines

**--branch**="": remove pipelines of this branch only

**--dry-run**: disable non-read api calls (default: false)

**--keep-min**="": minimum number of pipelines to keep (default: 10)

**--older-than**="": remove pipelines older than the specified time limit (default: 0s)

### queue

show pipeline queue

**--format**="": format output (default: \x1b[33m{{ .FullName }} #{{ .Number }} \x1b[0m\nStatus: {{ .Status }}\nEvent: {{ .Event }}\nCommit: {{ .Commit }}\nBranch: {{ .Branch }}\nRef: {{ .Ref }}\nAuthor: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}\nMessage: {{ .Message }}\n)

### show

show pipeline information

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers (default: false)

### start

start a pipeline

**--param, -p**="": custom parameters to inject into the step environment. Format: KEY=value

### stop

stop a pipeline

## repo

manage repositories

### add

add a repository

### chown

assume ownership of a repository

### cron

manage cron jobs

#### add

add a cron job

**--branch**="": cron branch

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nID: {{ .ID }}\nBranch: {{ .Branch }}\nSchedule: {{ .Schedule }}\nNextExec: {{ .NextExec }}\n)

**--name**="": cron name

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

#### rm

remove a cron job

**--id**="": cron id

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### ls

list cron jobs

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nID: {{ .ID }}\nBranch: {{ .Branch }}\nSchedule: {{ .Schedule }}\nNextExec: {{ .NextExec }}\n)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### show

show cron job information

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nID: {{ .ID }}\nBranch: {{ .Branch }}\nSchedule: {{ .Schedule }}\nNextExec: {{ .NextExec }}\n)

**--id**="": cron id

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### update

update a cron job

**--branch**="": cron branch

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nID: {{ .ID }}\nBranch: {{ .Branch }}\nSchedule: {{ .Schedule }}\nNextExec: {{ .NextExec }}\n)

**--id**="": cron id

**--name**="": cron name

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--schedule**="": cron schedule

### ls

list all repos

**--all**: query all repos, including inactive ones (default: false)

**--format**="": format output (deprecated)

**--org**="": filter by organization

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers (default: false)

### registry

manage registries

#### add

add a registry

**--hostname**="": registry hostname (default: docker.io)

**--password**="": registry password

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--username**="": registry username

#### rm

remove a registry

**--hostname**="": registry hostname (default: docker.io)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### ls

list registries

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Address }} \x1b[0m\nUsername: {{ .Username }}\nEmail: {{ .Email }}\n)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### show

show registry information

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Address }} \x1b[0m\nUsername: {{ .Username }}\nEmail: {{ .Email }}\n)

**--hostname**="": registry hostname (default: docker.io)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### update

update a registry

**--hostname**="": registry hostname (default: docker.io)

**--password**="": registry password

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--username**="": registry username

### rm

remove a repository

### repair

repair repository webhooks

### secret

manage secrets

#### add

add a secret

**--event**="": limit secret to these events

**--image**="": limit secret to these images

**--name**="": secret name

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--value**="": secret value

#### rm

remove a secret

**--name**="": secret name

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### ls

list secrets

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nEvents: {{ list .Events }}\n{{- if .Images }}\nImages: {{ list .Images }}\n{{- else }}\nImages: <any>\n{{- end }}\n)

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### show

show secret information

**--format**="": format output (deprecated) (default: \x1b[33m{{ .Name }} \x1b[0m\nEvents: {{ list .Events }}\n{{- if .Images }}\nImages: {{ list .Images }}\n{{- else }}\nImages: <any>\n{{- end }}\n)

**--name**="": secret name

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

#### update

update a secret

**--event**="": limit secret to these events

**--image**="": limit secret to these images

**--name**="": secret name

**--repository, --repo**="": repository id or full name (e.g. 134 or octocat/hello-world)

**--value**="": secret value

### show

show repository information

**--output**="": output format (default: table)

**--output-no-headers**: don't print headers (default: false)

### sync

synchronize the repository list

**--format**="": format output (default: \x1b[33m{{ .FullName }}\x1b[0m (id: {{ .ID }}, forgeRemoteID: {{ .ForgeRemoteID }}, isActive: {{ .IsActive }}))

### update

update a repository

**--config**="": repository configuration path. Example: .woodpecker.yml

**--pipeline-counter**="": repository starting pipeline number (default: 0)

**--require-approval**="": repository requires approval for

**--timeout**="": repository timeout (default: 0s)

**--trusted-network**: repository is network trusted (default: false)

**--trusted-security**: repository is security trusted (default: false)

**--trusted-volumes**: repository is volumes trusted (default: false)

**--unsafe**: allow unsafe operations (default: false)

**--visibility**="": repository visibility

## setup

setup the woodpecker-cli for the first time

**--server**="": URL of the woodpecker server

**--token**="": token to authenticate with the woodpecker server

## update

update the woodpecker-cli to the latest version

**--force**: force update even if the latest version is already installed (default: false)
