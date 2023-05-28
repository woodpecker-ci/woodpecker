
# REST API

Woodpecker offers a comprehensive REST API, so you can integrate easily with from and with other tools.

## API specification

Starting with Woodpecker v0.15.10+ a Swagger v2 API specification is served by the Woodpecker Server.
The typical URL looks like "http://woodpecker-host/swagger/doc.json", where you can fetch the API specification.

## Swagger API UI

Starting with Woodpecker v0.15.10+ a Swagger web user interface (UI) is served by the Woodpecker Server.
Typically, you can open "http://woodpecker-host/swagger/index.html" in your browser, to explore the API documentation.

# API endpoint summary

This is a summary of available API endpoints.
Please, keep in mind this documentation reflects latest development changes
and might differ from your used server version.
Its recommended to consult the Swagger API UI of your Woodpecker server,
where you also have the chance to do manual exploration and live testing.

## All endpoints

###  agents

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/agents/{agent} | [delete agents agent](#delete-agents-agent) | Delete an agent |
| GET | /api/agents | [get agents](#get-agents) | Get agent list |
| GET | /api/agents/{agent} | [get agents agent](#get-agents-agent) | Get agent information |
| GET | /api/agents/{agent}/tasks | [get agents agent tasks](#get-agents-agent-tasks) | Get agent tasks |
| PATCH | /api/agents/{agent} | [patch agents agent](#patch-agents-agent) | Update agent information |
| POST | /api/agents | [post agents](#post-agents) | Create a new agent with a random token so a new agent can connect to the server |
  


###  badges

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/badges/{owner}/{name}/cc.xml | [get badges owner name cc XML](#get-badges-owner-name-cc-xml) | Provide pipeline status information to the CCMenu tool |
| GET | /api/badges/{owner}/{name}/status.svg | [get badges owner name status svg](#get-badges-owner-name-status-svg) | Get status badge, SVG format |
  


###  organization_permissions

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/orgs/{owner}/permissions | [get orgs owner permissions](#get-orgs-owner-permissions) | Get the permissions of the current user in the given organization |
  


###  organization_secrets

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/orgs/{owner}/secrets/{secret} | [delete orgs owner secrets secret](#delete-orgs-owner-secrets-secret) | Delete the named secret from an organization |
| GET | /api/orgs/{owner}/secrets | [get orgs owner secrets](#get-orgs-owner-secrets) | Get the organization secret list |
| GET | /api/orgs/{owner}/secrets/{secret} | [get orgs owner secrets secret](#get-orgs-owner-secrets-secret) | Get the named organization secret |
| PATCH | /api/orgs/{owner}/secrets/{secret} | [patch orgs owner secrets secret](#patch-orgs-owner-secrets-secret) | Update an organization secret |
| POST | /api/orgs/{owner}/secrets | [post orgs owner secrets](#post-orgs-owner-secrets) | Persist/create an organization secret |
  


###  pipeline_files

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/repos/{owner}/{name}/files/{number} | [get repos owner name files number](#get-repos-owner-name-files-number) | Gets a list file by pipeline |
| GET | /api/repos/{owner}/{name}/files/{number}/{step}/{file} | [get repos owner name files number step file](#get-repos-owner-name-files-number-step-file) | Gets a file by process and name |
  


###  pipeline_logs

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/repos/{owner}/{name}/logs/{number}/{pid} | [get repos owner name logs number pid](#get-repos-owner-name-logs-number-pid) | Log information |
| GET | /api/repos/{owner}/{name}/logs/{number}/{pid}/{step} | [get repos owner name logs number pid step](#get-repos-owner-name-logs-number-pid-step) | Log information per step |
| POST | /api/repos/{owner}/{name}/logs/{number} | [post repos owner name logs number](#post-repos-owner-name-logs-number) | Deletes log |
  


###  pipeline_queues

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/pipelines | [get pipelines](#get-pipelines) | List pipeline queues |
| GET | /api/queue/info | [get queue info](#get-queue-info) | Get pipeline queue information |
| GET | /api/queue/norunningpipelines | [get queue norunningpipelines](#get-queue-norunningpipelines) | Block til pipeline queue has a running item |
| POST | /api/queue/pause | [post queue pause](#post-queue-pause) | Pause a pipeline queue |
| POST | /api/queue/resume | [post queue resume](#post-queue-resume) | Resume a pipeline queue |
  


###  pipelines

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/repos/{owner}/{name}/pipelines | [get repos owner name pipelines](#get-repos-owner-name-pipelines) | Get pipelines, current running and past ones |
| GET | /api/repos/{owner}/{name}/pipelines/{number} | [get repos owner name pipelines number](#get-repos-owner-name-pipelines-number) | Pipeline information by number |
| GET | /api/repos/{owner}/{name}/pipelines/{number}/config | [get repos owner name pipelines number config](#get-repos-owner-name-pipelines-number-config) | Pipeline configuration |
| POST | /api/repos/{owner}/{name}/pipelines | [post repos owner name pipelines](#post-repos-owner-name-pipelines) | Run/trigger a pipelines |
| POST | /api/repos/{owner}/{name}/pipelines/{number} | [post repos owner name pipelines number](#post-repos-owner-name-pipelines-number) | Restart a pipeline |
| POST | /api/repos/{owner}/{name}/pipelines/{number}/approve | [post repos owner name pipelines number approve](#post-repos-owner-name-pipelines-number-approve) | Start pipelines in gated repos |
| POST | /api/repos/{owner}/{name}/pipelines/{number}/cancel | [post repos owner name pipelines number cancel](#post-repos-owner-name-pipelines-number-cancel) | Cancels a pipeline |
| POST | /api/repos/{owner}/{name}/pipelines/{number}/decline | [post repos owner name pipelines number decline](#post-repos-owner-name-pipelines-number-decline) | Decline pipelines in gated repos |
  


###  process_profiling_and_debugging

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/debug/pprof | [get debug pprof](#get-debug-pprof) | List available pprof profiles (HTML) |
| GET | /api/debug/pprof/block | [get debug pprof block](#get-debug-pprof-block) | Get pprof stack traces that led to blocking on synchronization primitives |
| GET | /api/debug/pprof/cmdline | [get debug pprof cmdline](#get-debug-pprof-cmdline) | Get the command line invocation of the current program |
| GET | /api/debug/pprof/goroutine | [get debug pprof goroutine](#get-debug-pprof-goroutine) | Get pprof stack traces of all current goroutines |
| GET | /api/debug/pprof/heap | [get debug pprof heap](#get-debug-pprof-heap) | Get pprof heap dump, a sampling of memory allocations of live objects |
| GET | /api/debug/pprof/profile | [get debug pprof profile](#get-debug-pprof-profile) | Get pprof CPU profile |
| GET | /api/debug/pprof/symbol | [get debug pprof symbol](#get-debug-pprof-symbol) | Get pprof program counters mapping to function names |
| GET | /api/debug/pprof/threadcreate | [get debug pprof threadcreate](#get-debug-pprof-threadcreate) | Get pprof stack traces that led to the creation of new OS threads |
| GET | /api/debug/pprof/trace | [get debug pprof trace](#get-debug-pprof-trace) | Get a trace of execution of the current program |
| POST | /api/debug/pprof/symbol | [post debug pprof symbol](#post-debug-pprof-symbol) | Get pprof program counters mapping to function names |
  


###  repositories

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/repos/{owner}/{name} | [delete repos owner name](#delete-repos-owner-name) | Delete a repository |
| GET | /api/repos/{owner}/{name} | [get repos owner name](#get-repos-owner-name) | Get repository information |
| GET | /api/repos/{owner}/{name}/branches | [get repos owner name branches](#get-repos-owner-name-branches) | Get repository branches |
| GET | /api/repos/{owner}/{name}/permissions | [get repos owner name permissions](#get-repos-owner-name-permissions) | Repository permission information |
| GET | /api/repos/{owner}/{name}/pull_requests | [get repos owner name pull requests](#get-repos-owner-name-pull-requests) | List active pull requests |
| PATCH | /api/repos/{owner}/{name} | [patch repos owner name](#patch-repos-owner-name) | Change a repository |
| POST | /api/repos/{owner}/{name} | [post repos owner name](#post-repos-owner-name) | Activate a repository |
| POST | /api/repos/{owner}/{name}/chown | [post repos owner name chown](#post-repos-owner-name-chown) | Change a repository's owner, to the one holding the access token |
| POST | /api/repos/{owner}/{name}/move | [post repos owner name move](#post-repos-owner-name-move) | Move a repository to a new owner |
| POST | /api/repos/{owner}/{name}/repair | [post repos owner name repair](#post-repos-owner-name-repair) | Repair a repository |
  


###  repository_cron_jobs

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/repos/{owner}/{name}/cron/{cron} | [delete repos owner name cron cron](#delete-repos-owner-name-cron-cron) | Delete a cron job by id |
| GET | /api/repos/{owner}/{name}/cron | [get repos owner name cron](#get-repos-owner-name-cron) | Get the cron job list |
| GET | /api/repos/{owner}/{name}/cron/{cron} | [get repos owner name cron cron](#get-repos-owner-name-cron-cron) | Get a cron job by id |
| PATCH | /api/repos/{owner}/{name}/cron/{cron} | [patch repos owner name cron cron](#patch-repos-owner-name-cron-cron) | Update a cron job |
| POST | /api/repos/{owner}/{name}/cron | [post repos owner name cron](#post-repos-owner-name-cron) | Persist/creat a cron job |
| POST | /api/repos/{owner}/{name}/cron/{cron} | [post repos owner name cron cron](#post-repos-owner-name-cron-cron) | Start a cron job now |
  


###  repository_registries

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/repos/{owner}/{name}/registry/{registry} | [delete repos owner name registry registry](#delete-repos-owner-name-registry-registry) | Delete a named registry |
| GET | /api/repos/{owner}/{name}/registry | [get repos owner name registry](#get-repos-owner-name-registry) | Get the registry list |
| GET | /api/repos/{owner}/{name}/registry/{registry} | [get repos owner name registry registry](#get-repos-owner-name-registry-registry) | Get a named registry |
| PATCH | /api/repos/{owner}/{name}/registry/{registry} | [patch repos owner name registry registry](#patch-repos-owner-name-registry-registry) | Update a named registry |
| POST | /api/repos/{owner}/{name}/registry | [post repos owner name registry](#post-repos-owner-name-registry) | Persist/create a registry |
  


###  repository_secrets

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/repos/{owner}/{name}/secrets/{secretName} | [delete repos owner name secrets secret name](#delete-repos-owner-name-secrets-secret-name) | Delete a named secret |
| GET | /api/repos/{owner}/{name}/secrets | [get repos owner name secrets](#get-repos-owner-name-secrets) | Get the secret list |
| GET | /api/repos/{owner}/{name}/secrets/{secretName} | [get repos owner name secrets secret name](#get-repos-owner-name-secrets-secret-name) | Get a named secret |
| PATCH | /api/repos/{owner}/{name}/secrets/{secretName} | [patch repos owner name secrets secret name](#patch-repos-owner-name-secrets-secret-name) | Update a named secret |
| POST | /api/repos/{owner}/{name}/secrets | [post repos owner name secrets](#post-repos-owner-name-secrets) | Persist/create a secret |
  


###  secrets

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/secrets/{secret} | [delete secrets secret](#delete-secrets-secret) | Delete a global secret by name |
| GET | /api/secrets | [get secrets](#get-secrets) | Get the global secret list |
| GET | /api/secrets/{secret} | [get secrets secret](#get-secrets-secret) | Get a global secret by name |
| PATCH | /api/secrets/{secret} | [patch secrets secret](#patch-secrets-secret) | Update a global secret by name |
| POST | /api/secrets | [post secrets](#post-secrets) | Persist/create a global secret |
  


###  system

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/healthz | [get healthz](#get-healthz) | Health information |
| GET | /api/log-level | [get log level](#get-log-level) | Current log level |
| GET | /api/signature/public-key | [get signature public key](#get-signature-public-key) | Get server's signature public key |
| GET | /api/version | [get version](#get-version) | Get version |
| POST | /api/hook | [post hook](#post-hook) | Incoming webhook from Github or Gitea |
| POST | /api/log-level | [post log level](#post-log-level) | Set log level |
  


###  user

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/user/token | [delete user token](#delete-user-token) | Reset a token |
| GET | /api/user | [get user](#get-user) | Returns the currently authenticated user. |
| GET | /api/user/feed | [get user feed](#get-user-feed) | A feed entry for a build. |
| GET | /api/user/repos | [get user repos](#get-user-repos) | Get user's repos |
| POST | /api/user/token | [post user token](#post-user-token) | tbd |
  


###  users

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /api/users/{login} | [delete users login](#delete-users-login) | Delete a user |
| GET | /api/users | [get users](#get-users) | Get all users |
| GET | /api/users/{login} | [get users login](#get-users-login) | Get a user |
| PATCH | /api/users/{login} | [patch users login](#patch-users-login) | Change a user |
| POST | /api/users | [post users](#post-users) | Create a user |
  


## Paths

### <span id="delete-agents-agent"></span> Delete an agent (*DeleteAgentsAgent*)

```
DELETE /api/agents/{agent}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| agent | `path` | integer | `int64` |  | ✓ |  | the agent's id |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-agents-agent-200) | OK | OK |  | [schema](#delete-agents-agent-200-schema) |

#### Responses


##### <span id="delete-agents-agent-200"></span> 200 - OK
Status: OK

###### <span id="delete-agents-agent-200-schema"></span> Schema

### <span id="delete-orgs-owner-secrets-secret"></span> Delete the named secret from an organization (*DeleteOrgsOwnerSecretsSecret*)

```
DELETE /api/orgs/{owner}/secrets/{secret}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| owner | `path` | string | `string` |  | ✓ |  | the owner's name |
| secret | `path` | string | `string` |  | ✓ |  | the secret's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-orgs-owner-secrets-secret-200) | OK | OK |  | [schema](#delete-orgs-owner-secrets-secret-200-schema) |

#### Responses


##### <span id="delete-orgs-owner-secrets-secret-200"></span> 200 - OK
Status: OK

###### <span id="delete-orgs-owner-secrets-secret-200-schema"></span> Schema

### <span id="delete-repos-owner-name"></span> Delete a repository (*DeleteReposOwnerName*)

```
DELETE /api/repos/{owner}/{name}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-repos-owner-name-200) | OK | OK |  | [schema](#delete-repos-owner-name-200-schema) |

#### Responses


##### <span id="delete-repos-owner-name-200"></span> 200 - OK
Status: OK

###### <span id="delete-repos-owner-name-200-schema"></span> Schema
   
  

[Repo](#repo)

### <span id="delete-repos-owner-name-cron-cron"></span> Delete a cron job by id (*DeleteReposOwnerNameCronCron*)

```
DELETE /api/repos/{owner}/{name}/cron/{cron}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| cron | `path` | string | `string` |  | ✓ |  | the cron job id |
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-repos-owner-name-cron-cron-200) | OK | OK |  | [schema](#delete-repos-owner-name-cron-cron-200-schema) |

#### Responses


##### <span id="delete-repos-owner-name-cron-cron-200"></span> 200 - OK
Status: OK

###### <span id="delete-repos-owner-name-cron-cron-200-schema"></span> Schema

### <span id="delete-repos-owner-name-registry-registry"></span> Delete a named registry (*DeleteReposOwnerNameRegistryRegistry*)

```
DELETE /api/repos/{owner}/{name}/registry/{registry}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| registry | `path` | string | `string` |  | ✓ |  | the registry name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-repos-owner-name-registry-registry-200) | OK | OK |  | [schema](#delete-repos-owner-name-registry-registry-200-schema) |

#### Responses


##### <span id="delete-repos-owner-name-registry-registry-200"></span> 200 - OK
Status: OK

###### <span id="delete-repos-owner-name-registry-registry-200-schema"></span> Schema

### <span id="delete-repos-owner-name-secrets-secret-name"></span> Delete a named secret (*DeleteReposOwnerNameSecretsSecretName*)

```
DELETE /api/repos/{owner}/{name}/secrets/{secretName}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| secretName | `path` | string | `string` |  | ✓ |  | the secret name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-repos-owner-name-secrets-secret-name-200) | OK | OK |  | [schema](#delete-repos-owner-name-secrets-secret-name-200-schema) |

#### Responses


##### <span id="delete-repos-owner-name-secrets-secret-name-200"></span> 200 - OK
Status: OK

###### <span id="delete-repos-owner-name-secrets-secret-name-200-schema"></span> Schema

### <span id="delete-secrets-secret"></span> Delete a global secret by name (*DeleteSecretsSecret*)

```
DELETE /api/secrets/{secret}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| secret | `path` | string | `string` |  | ✓ |  | the secret's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-secrets-secret-200) | OK | OK |  | [schema](#delete-secrets-secret-200-schema) |

#### Responses


##### <span id="delete-secrets-secret-200"></span> 200 - OK
Status: OK

###### <span id="delete-secrets-secret-200-schema"></span> Schema

### <span id="delete-user-token"></span> Reset a token (*DeleteUserToken*)

```
DELETE /api/user/token
```

Reset's the current personal access token of the user and returns a new one.

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-user-token-200) | OK | OK |  | [schema](#delete-user-token-200-schema) |

#### Responses


##### <span id="delete-user-token-200"></span> 200 - OK
Status: OK

###### <span id="delete-user-token-200-schema"></span> Schema

### <span id="delete-users-login"></span> Delete a user (*DeleteUsersLogin*)

```
DELETE /api/users/{login}
```

Deletes the given user. Requires admin rights.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| login | `path` | string | `string` |  | ✓ |  | the user's login name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#delete-users-login-200) | OK | OK |  | [schema](#delete-users-login-200-schema) |

#### Responses


##### <span id="delete-users-login-200"></span> 200 - OK
Status: OK

###### <span id="delete-users-login-200-schema"></span> Schema
   
  

[User](#user)

### <span id="get-agents"></span> Get agent list (*GetAgents*)

```
GET /api/agents
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-agents-200) | OK | OK |  | [schema](#get-agents-200-schema) |

#### Responses


##### <span id="get-agents-200"></span> 200 - OK
Status: OK

###### <span id="get-agents-200-schema"></span> Schema
   
  

[][Agent](#agent)

### <span id="get-agents-agent"></span> Get agent information (*GetAgentsAgent*)

```
GET /api/agents/{agent}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| agent | `path` | integer | `int64` |  | ✓ |  | the agent's id |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-agents-agent-200) | OK | OK |  | [schema](#get-agents-agent-200-schema) |

#### Responses


##### <span id="get-agents-agent-200"></span> 200 - OK
Status: OK

###### <span id="get-agents-agent-200-schema"></span> Schema
   
  

[Agent](#agent)

### <span id="get-agents-agent-tasks"></span> Get agent tasks (*GetAgentsAgentTasks*)

```
GET /api/agents/{agent}/tasks
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| agent | `path` | integer | `int64` |  | ✓ |  | the agent's id |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-agents-agent-tasks-200) | OK | OK |  | [schema](#get-agents-agent-tasks-200-schema) |

#### Responses


##### <span id="get-agents-agent-tasks-200"></span> 200 - OK
Status: OK

###### <span id="get-agents-agent-tasks-200-schema"></span> Schema
   
  

[][Task](#task)

### <span id="get-badges-owner-name-cc-xml"></span> Provide pipeline status information to the CCMenu tool (*GetBadgesOwnerNameCcXML*)

```
GET /api/badges/{owner}/{name}/cc.xml
```

CCMenu displays the pipeline status of projects on a CI server as an item in the Mac's menu bar.
It started as part of the CruiseControl project that built the first continuous integration server.
More details on how to install, you can find at http://ccmenu.org/
The response format adheres to CCTray v1 Specification, https://cctray.org/v1/

#### Produces
  * text/xml

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-badges-owner-name-cc-xml-200) | OK | OK |  | [schema](#get-badges-owner-name-cc-xml-200-schema) |

#### Responses


##### <span id="get-badges-owner-name-cc-xml-200"></span> 200 - OK
Status: OK

###### <span id="get-badges-owner-name-cc-xml-200-schema"></span> Schema

### <span id="get-badges-owner-name-status-svg"></span> Get status badge, SVG format (*GetBadgesOwnerNameStatusSvg*)

```
GET /api/badges/{owner}/{name}/status.svg
```

#### Produces
  * image/svg+xml

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-badges-owner-name-status-svg-200) | OK | OK |  | [schema](#get-badges-owner-name-status-svg-200-schema) |

#### Responses


##### <span id="get-badges-owner-name-status-svg-200"></span> 200 - OK
Status: OK

###### <span id="get-badges-owner-name-status-svg-200-schema"></span> Schema

### <span id="get-debug-pprof"></span> List available pprof profiles (HTML) (*GetDebugPprof*)

```
GET /api/debug/pprof
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug

#### Produces
  * text/html

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-200) | OK | OK |  | [schema](#get-debug-pprof-200-schema) |

#### Responses


##### <span id="get-debug-pprof-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-200-schema"></span> Schema

### <span id="get-debug-pprof-block"></span> Get pprof stack traces that led to blocking on synchronization primitives (*GetDebugPprofBlock*)

```
GET /api/debug/pprof/block
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-block-200) | OK | OK |  | [schema](#get-debug-pprof-block-200-schema) |

#### Responses


##### <span id="get-debug-pprof-block-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-block-200-schema"></span> Schema

### <span id="get-debug-pprof-cmdline"></span> Get the command line invocation of the current program (*GetDebugPprofCmdline*)

```
GET /api/debug/pprof/cmdline
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-cmdline-200) | OK | OK |  | [schema](#get-debug-pprof-cmdline-200-schema) |

#### Responses


##### <span id="get-debug-pprof-cmdline-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-cmdline-200-schema"></span> Schema

### <span id="get-debug-pprof-goroutine"></span> Get pprof stack traces of all current goroutines (*GetDebugPprofGoroutine*)

```
GET /api/debug/pprof/goroutine
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| debug | `query` | integer | `int64` |  |  | `1` | Use debug=2 as a query parameter to export in the same format as an un-recovered panic |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-goroutine-200) | OK | OK |  | [schema](#get-debug-pprof-goroutine-200-schema) |

#### Responses


##### <span id="get-debug-pprof-goroutine-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-goroutine-200-schema"></span> Schema

### <span id="get-debug-pprof-heap"></span> Get pprof heap dump, a sampling of memory allocations of live objects (*GetDebugPprofHeap*)

```
GET /api/debug/pprof/heap
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| gc | `query` | string | `string` |  |  |  | You can specify gc=heap to run GC before taking the heap sample |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-heap-200) | OK | OK |  | [schema](#get-debug-pprof-heap-200-schema) |

#### Responses


##### <span id="get-debug-pprof-heap-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-heap-200-schema"></span> Schema

### <span id="get-debug-pprof-profile"></span> Get pprof CPU profile (*GetDebugPprofProfile*)

```
GET /api/debug/pprof/profile
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
After you get the profile file, use the go tool pprof command to investigate the profile.

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| seconds | `query` | integer | `int64` |  | ✓ |  | You can specify the duration in the seconds GET parameter. |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-profile-200) | OK | OK |  | [schema](#get-debug-pprof-profile-200-schema) |

#### Responses


##### <span id="get-debug-pprof-profile-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-profile-200-schema"></span> Schema

### <span id="get-debug-pprof-symbol"></span> Get pprof program counters mapping to function names (*GetDebugPprofSymbol*)

```
GET /api/debug/pprof/symbol
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
Looks up the program counters listed in the request,
responding with a table mapping program counters to function names.
The requested program counters can be provided via GET + query parameters,
or POST + body parameters. Program counters shall be space delimited.

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-symbol-200) | OK | OK |  | [schema](#get-debug-pprof-symbol-200-schema) |

#### Responses


##### <span id="get-debug-pprof-symbol-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-symbol-200-schema"></span> Schema

### <span id="get-debug-pprof-threadcreate"></span> Get pprof stack traces that led to the creation of new OS threads (*GetDebugPprofThreadcreate*)

```
GET /api/debug/pprof/threadcreate
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-threadcreate-200) | OK | OK |  | [schema](#get-debug-pprof-threadcreate-200-schema) |

#### Responses


##### <span id="get-debug-pprof-threadcreate-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-threadcreate-200-schema"></span> Schema

### <span id="get-debug-pprof-trace"></span> Get a trace of execution of the current program (*GetDebugPprofTrace*)

```
GET /api/debug/pprof/trace
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
After you get the profile file, use the go tool pprof command to investigate the profile.

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| seconds | `query` | integer | `int64` |  | ✓ |  | You can specify the duration in the seconds GET parameter. |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-debug-pprof-trace-200) | OK | OK |  | [schema](#get-debug-pprof-trace-200-schema) |

#### Responses


##### <span id="get-debug-pprof-trace-200"></span> 200 - OK
Status: OK

###### <span id="get-debug-pprof-trace-200-schema"></span> Schema

### <span id="get-healthz"></span> Health information (*GetHealthz*)

```
GET /api/healthz
```

If everything is fine, just a 200 will be returned, a 500 signals server state is unhealthy.

#### Produces
  * text/plain

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-healthz-200) | OK | OK |  | [schema](#get-healthz-200-schema) |
| [500](#get-healthz-500) | Internal Server Error | Internal Server Error |  | [schema](#get-healthz-500-schema) |

#### Responses


##### <span id="get-healthz-200"></span> 200 - OK
Status: OK

###### <span id="get-healthz-200-schema"></span> Schema

##### <span id="get-healthz-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-healthz-500-schema"></span> Schema

### <span id="get-log-level"></span> Current log level (*GetLogLevel*)

```
GET /api/log-level
```

Endpoint returns the current logging level. Requires admin rights.

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-log-level-200) | OK | OK |  | [schema](#get-log-level-200-schema) |

#### Responses


##### <span id="get-log-level-200"></span> 200 - OK
Status: OK

###### <span id="get-log-level-200-schema"></span> Schema
   
  

[GetLogLevelOKBody](#get-log-level-o-k-body)

###### Inlined models

**<span id="get-log-level-o-k-body"></span> GetLogLevelOKBody**


  


* composed type [GetLogLevelOKBodyAllOf0](#get-log-level-o-k-body-all-of0)
* inlined member (*getLogLevelOKBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| log-level | string| `string` |  | |  |  |



**<span id="get-log-level-o-k-body-all-of0"></span> GetLogLevelOKBodyAllOf0**


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| GetLogLevelOKBodyAllOf0 | string| string | |  |  |



### <span id="get-orgs-owner-permissions"></span> Get the permissions of the current user in the given organization (*GetOrgsOwnerPermissions*)

```
GET /api/orgs/{owner}/permissions
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| owner | `path` | string | `string` |  | ✓ |  | the owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-orgs-owner-permissions-200) | OK | OK |  | [schema](#get-orgs-owner-permissions-200-schema) |

#### Responses


##### <span id="get-orgs-owner-permissions-200"></span> 200 - OK
Status: OK

###### <span id="get-orgs-owner-permissions-200-schema"></span> Schema
   
  

[][OrgPerm](#org-perm)

### <span id="get-orgs-owner-secrets"></span> Get the organization secret list (*GetOrgsOwnerSecrets*)

```
GET /api/orgs/{owner}/secrets
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| owner | `path` | string | `string` |  | ✓ |  | the owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-orgs-owner-secrets-200) | OK | OK |  | [schema](#get-orgs-owner-secrets-200-schema) |

#### Responses


##### <span id="get-orgs-owner-secrets-200"></span> 200 - OK
Status: OK

###### <span id="get-orgs-owner-secrets-200-schema"></span> Schema
   
  

[][Secret](#secret)

### <span id="get-orgs-owner-secrets-secret"></span> Get the named organization secret (*GetOrgsOwnerSecretsSecret*)

```
GET /api/orgs/{owner}/secrets/{secret}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| owner | `path` | string | `string` |  | ✓ |  | the owner's name |
| secret | `path` | string | `string` |  | ✓ |  | the secret's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-orgs-owner-secrets-secret-200) | OK | OK |  | [schema](#get-orgs-owner-secrets-secret-200-schema) |

#### Responses


##### <span id="get-orgs-owner-secrets-secret-200"></span> 200 - OK
Status: OK

###### <span id="get-orgs-owner-secrets-secret-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="get-pipelines"></span> List pipeline queues (*GetPipelines*)

```
GET /api/pipelines
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-pipelines-200) | OK | OK |  | [schema](#get-pipelines-200-schema) |

#### Responses


##### <span id="get-pipelines-200"></span> 200 - OK
Status: OK

###### <span id="get-pipelines-200-schema"></span> Schema
   
  

[][Feed](#feed)

### <span id="get-queue-info"></span> Get pipeline queue information (*GetQueueInfo*)

```
GET /api/queue/info
```

TODO: link the InfoT response object - this is blocked, until this swag-issue is solved ...

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-queue-info-200) | OK | OK |  | [schema](#get-queue-info-200-schema) |

#### Responses


##### <span id="get-queue-info-200"></span> 200 - OK
Status: OK

###### <span id="get-queue-info-200-schema"></span> Schema
   
  

map of string

### <span id="get-queue-norunningpipelines"></span> Block til pipeline queue has a running item (*GetQueueNorunningpipelines*)

```
GET /api/queue/norunningpipelines
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-queue-norunningpipelines-200) | OK | OK |  | [schema](#get-queue-norunningpipelines-200-schema) |

#### Responses


##### <span id="get-queue-norunningpipelines-200"></span> 200 - OK
Status: OK

###### <span id="get-queue-norunningpipelines-200-schema"></span> Schema

### <span id="get-repos-owner-name"></span> Get repository information (*GetReposOwnerName*)

```
GET /api/repos/{owner}/{name}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-200) | OK | OK |  | [schema](#get-repos-owner-name-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-200-schema"></span> Schema
   
  

[Repo](#repo)

### <span id="get-repos-owner-name-branches"></span> Get repository branches (*GetReposOwnerNameBranches*)

```
GET /api/repos/{owner}/{name}/branches
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-branches-200) | OK | OK |  | [schema](#get-repos-owner-name-branches-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-branches-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-branches-200-schema"></span> Schema
   
  

[]string

### <span id="get-repos-owner-name-cron"></span> Get the cron job list (*GetReposOwnerNameCron*)

```
GET /api/repos/{owner}/{name}/cron
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-cron-200) | OK | OK |  | [schema](#get-repos-owner-name-cron-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-cron-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-cron-200-schema"></span> Schema
   
  

[][Cron](#cron)

### <span id="get-repos-owner-name-cron-cron"></span> Get a cron job by id (*GetReposOwnerNameCronCron*)

```
GET /api/repos/{owner}/{name}/cron/{cron}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| cron | `path` | string | `string` |  | ✓ |  | the cron job id |
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-cron-cron-200) | OK | OK |  | [schema](#get-repos-owner-name-cron-cron-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-cron-cron-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-cron-cron-200-schema"></span> Schema
   
  

[Cron](#cron)

### <span id="get-repos-owner-name-files-number"></span> Gets a list file by pipeline (*GetReposOwnerNameFilesNumber*)

```
GET /api/repos/{owner}/{name}/files/{number}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-files-number-200) | OK | OK |  | [schema](#get-repos-owner-name-files-number-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-files-number-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-files-number-200-schema"></span> Schema
   
  

[][File](#file)

### <span id="get-repos-owner-name-files-number-step-file"></span> Gets a file by process and name (*GetReposOwnerNameFilesNumberStepFile*)

```
GET /api/repos/{owner}/{name}/files/{number}/{step}/{file}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| file | `path` | string | `string` |  | ✓ |  | the filename |
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| step | `path` | integer | `int64` |  | ✓ |  | the step of the pipeline |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-files-number-step-file-200) | OK | OK |  | [schema](#get-repos-owner-name-files-number-step-file-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-files-number-step-file-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-files-number-step-file-200-schema"></span> Schema

### <span id="get-repos-owner-name-logs-number-pid"></span> Log information (*GetReposOwnerNameLogsNumberPid*)

```
GET /api/repos/{owner}/{name}/logs/{number}/{pid}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| pid | `path` | integer | `int64` |  | ✓ |  | the pipeline id |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-logs-number-pid-200) | OK | OK |  | [schema](#get-repos-owner-name-logs-number-pid-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-logs-number-pid-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-logs-number-pid-200-schema"></span> Schema

### <span id="get-repos-owner-name-logs-number-pid-step"></span> Log information per step (*GetReposOwnerNameLogsNumberPidStep*)

```
GET /api/repos/{owner}/{name}/logs/{number}/{pid}/{step}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| pid | `path` | integer | `int64` |  | ✓ |  | the pipeline id |
| step | `path` | integer | `int64` |  | ✓ |  | the step name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-logs-number-pid-step-200) | OK | OK |  | [schema](#get-repos-owner-name-logs-number-pid-step-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-logs-number-pid-step-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-logs-number-pid-step-200-schema"></span> Schema

### <span id="get-repos-owner-name-permissions"></span> Repository permission information (*GetReposOwnerNamePermissions*)

```
GET /api/repos/{owner}/{name}/permissions
```

The repository permission, according to the used access token.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-permissions-200) | OK | OK |  | [schema](#get-repos-owner-name-permissions-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-permissions-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-permissions-200-schema"></span> Schema
   
  

[Perm](#perm)

### <span id="get-repos-owner-name-pipelines"></span> Get pipelines, current running and past ones (*GetReposOwnerNamePipelines*)

```
GET /api/repos/{owner}/{name}/pipelines
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-pipelines-200) | OK | OK |  | [schema](#get-repos-owner-name-pipelines-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-pipelines-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-pipelines-200-schema"></span> Schema
   
  

[][Pipeline](#pipeline)

### <span id="get-repos-owner-name-pipelines-number"></span> Pipeline information by number (*GetReposOwnerNamePipelinesNumber*)

```
GET /api/repos/{owner}/{name}/pipelines/{number}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline, OR 'latest' |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-pipelines-number-200) | OK | OK |  | [schema](#get-repos-owner-name-pipelines-number-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-pipelines-number-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-pipelines-number-200-schema"></span> Schema
   
  

[Pipeline](#pipeline)

### <span id="get-repos-owner-name-pipelines-number-config"></span> Pipeline configuration (*GetReposOwnerNamePipelinesNumberConfig*)

```
GET /api/repos/{owner}/{name}/pipelines/{number}/config
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-pipelines-number-config-200) | OK | OK |  | [schema](#get-repos-owner-name-pipelines-number-config-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-pipelines-number-config-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-pipelines-number-config-200-schema"></span> Schema
   
  

[][Config](#config)

### <span id="get-repos-owner-name-pull-requests"></span> List active pull requests (*GetReposOwnerNamePullRequests*)

```
GET /api/repos/{owner}/{name}/pull_requests
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-pull-requests-200) | OK | OK |  | [schema](#get-repos-owner-name-pull-requests-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-pull-requests-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-pull-requests-200-schema"></span> Schema
   
  

[][PullRequest](#pull-request)

### <span id="get-repos-owner-name-registry"></span> Get the registry list (*GetReposOwnerNameRegistry*)

```
GET /api/repos/{owner}/{name}/registry
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-registry-200) | OK | OK |  | [schema](#get-repos-owner-name-registry-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-registry-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-registry-200-schema"></span> Schema
   
  

[][Registry](#registry)

### <span id="get-repos-owner-name-registry-registry"></span> Get a named registry (*GetReposOwnerNameRegistryRegistry*)

```
GET /api/repos/{owner}/{name}/registry/{registry}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| registry | `path` | string | `string` |  | ✓ |  | the registry name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-registry-registry-200) | OK | OK |  | [schema](#get-repos-owner-name-registry-registry-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-registry-registry-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-registry-registry-200-schema"></span> Schema
   
  

[Registry](#registry)

### <span id="get-repos-owner-name-secrets"></span> Get the secret list (*GetReposOwnerNameSecrets*)

```
GET /api/repos/{owner}/{name}/secrets
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-secrets-200) | OK | OK |  | [schema](#get-repos-owner-name-secrets-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-secrets-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-secrets-200-schema"></span> Schema
   
  

[][Secret](#secret)

### <span id="get-repos-owner-name-secrets-secret-name"></span> Get a named secret (*GetReposOwnerNameSecretsSecretName*)

```
GET /api/repos/{owner}/{name}/secrets/{secretName}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| secretName | `path` | string | `string` |  | ✓ |  | the secret name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-repos-owner-name-secrets-secret-name-200) | OK | OK |  | [schema](#get-repos-owner-name-secrets-secret-name-200-schema) |

#### Responses


##### <span id="get-repos-owner-name-secrets-secret-name-200"></span> 200 - OK
Status: OK

###### <span id="get-repos-owner-name-secrets-secret-name-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="get-secrets"></span> Get the global secret list (*GetSecrets*)

```
GET /api/secrets
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-secrets-200) | OK | OK |  | [schema](#get-secrets-200-schema) |

#### Responses


##### <span id="get-secrets-200"></span> 200 - OK
Status: OK

###### <span id="get-secrets-200-schema"></span> Schema
   
  

[][Secret](#secret)

### <span id="get-secrets-secret"></span> Get a global secret by name (*GetSecretsSecret*)

```
GET /api/secrets/{secret}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| secret | `path` | string | `string` |  | ✓ |  | the secret's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-secrets-secret-200) | OK | OK |  | [schema](#get-secrets-secret-200-schema) |

#### Responses


##### <span id="get-secrets-secret-200"></span> 200 - OK
Status: OK

###### <span id="get-secrets-secret-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="get-signature-public-key"></span> Get server's signature public key (*GetSignaturePublicKey*)

```
GET /api/signature/public-key
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-signature-public-key-200) | OK | OK |  | [schema](#get-signature-public-key-200-schema) |

#### Responses


##### <span id="get-signature-public-key-200"></span> 200 - OK
Status: OK

###### <span id="get-signature-public-key-200-schema"></span> Schema

### <span id="get-user"></span> Returns the currently authenticated user. (*GetUser*)

```
GET /api/user
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-user-200) | OK | OK |  | [schema](#get-user-200-schema) |

#### Responses


##### <span id="get-user-200"></span> 200 - OK
Status: OK

###### <span id="get-user-200-schema"></span> Schema
   
  

[User](#user)

### <span id="get-user-feed"></span> A feed entry for a build. (*GetUserFeed*)

```
GET /api/user/feed
```

Feed entries can be used to display information on the latest builds.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-user-feed-200) | OK | OK |  | [schema](#get-user-feed-200-schema) |

#### Responses


##### <span id="get-user-feed-200"></span> 200 - OK
Status: OK

###### <span id="get-user-feed-200-schema"></span> Schema
   
  

[Feed](#feed)

### <span id="get-user-repos"></span> Get user's repos (*GetUserRepos*)

```
GET /api/user/repos
```

Retrieve the currently authenticated User's Repository list

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-user-repos-200) | OK | OK |  | [schema](#get-user-repos-200-schema) |

#### Responses


##### <span id="get-user-repos-200"></span> 200 - OK
Status: OK

###### <span id="get-user-repos-200-schema"></span> Schema
   
  

[][Repo](#repo)

### <span id="get-users"></span> Get all users (*GetUsers*)

```
GET /api/users
```

Returns all registered, active users in the system. Requires admin rights.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| page | `query` | integer | `int64` |  |  | `1` | for response pagination, page offset number |
| perPage | `query` | integer | `int64` |  |  | `50` | for response pagination, max items per page |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-users-200) | OK | OK |  | [schema](#get-users-200-schema) |

#### Responses


##### <span id="get-users-200"></span> 200 - OK
Status: OK

###### <span id="get-users-200-schema"></span> Schema
   
  

[][User](#user)

### <span id="get-users-login"></span> Get a user (*GetUsersLogin*)

```
GET /api/users/{login}
```

Returns a user with the specified login name. Requires admin rights.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| login | `path` | string | `string` |  | ✓ |  | the user's login name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-users-login-200) | OK | OK |  | [schema](#get-users-login-200-schema) |

#### Responses


##### <span id="get-users-login-200"></span> 200 - OK
Status: OK

###### <span id="get-users-login-200-schema"></span> Schema
   
  

[User](#user)

### <span id="get-version"></span> Get version (*GetVersion*)

```
GET /api/version
```

Endpoint returns the server version and build information.

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-version-200) | OK | OK |  | [schema](#get-version-200-schema) |

#### Responses


##### <span id="get-version-200"></span> 200 - OK
Status: OK

###### <span id="get-version-200-schema"></span> Schema
   
  

[GetVersionOKBody](#get-version-o-k-body)

###### Inlined models

**<span id="get-version-o-k-body"></span> GetVersionOKBody**


  


* composed type [GetVersionOKBodyAllOf0](#get-version-o-k-body-all-of0)
* inlined member (*getVersionOKBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| source | string| `string` |  | |  |  |
| version | string| `string` |  | |  |  |



**<span id="get-version-o-k-body-all-of0"></span> GetVersionOKBodyAllOf0**


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| GetVersionOKBodyAllOf0 | string| string | |  |  |



### <span id="patch-agents-agent"></span> Update agent information (*PatchAgentsAgent*)

```
PATCH /api/agents/{agent}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| agent | `path` | integer | `int64` |  | ✓ |  | the agent's id |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| agentData | `body` | [Agent](#agent) | `models.Agent` | | ✓ | | the agent's data |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-agents-agent-200) | OK | OK |  | [schema](#patch-agents-agent-200-schema) |

#### Responses


##### <span id="patch-agents-agent-200"></span> 200 - OK
Status: OK

###### <span id="patch-agents-agent-200-schema"></span> Schema
   
  

[Agent](#agent)

### <span id="patch-orgs-owner-secrets-secret"></span> Update an organization secret (*PatchOrgsOwnerSecretsSecret*)

```
PATCH /api/orgs/{owner}/secrets/{secret}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| owner | `path` | string | `string` |  | ✓ |  | the owner's name |
| secret | `path` | string | `string` |  | ✓ |  | the secret's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| secretData | `body` | [Secret](#secret) | `models.Secret` | | ✓ | | the update secret data |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-orgs-owner-secrets-secret-200) | OK | OK |  | [schema](#patch-orgs-owner-secrets-secret-200-schema) |

#### Responses


##### <span id="patch-orgs-owner-secrets-secret-200"></span> 200 - OK
Status: OK

###### <span id="patch-orgs-owner-secrets-secret-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="patch-repos-owner-name"></span> Change a repository (*PatchReposOwnerName*)

```
PATCH /api/repos/{owner}/{name}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| repo | `body` | [RepoPatch](#repo-patch) | `models.RepoPatch` | | ✓ | | the repository's information |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-repos-owner-name-200) | OK | OK |  | [schema](#patch-repos-owner-name-200-schema) |

#### Responses


##### <span id="patch-repos-owner-name-200"></span> 200 - OK
Status: OK

###### <span id="patch-repos-owner-name-200-schema"></span> Schema
   
  

[Repo](#repo)

### <span id="patch-repos-owner-name-cron-cron"></span> Update a cron job (*PatchReposOwnerNameCronCron*)

```
PATCH /api/repos/{owner}/{name}/cron/{cron}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| cron | `path` | string | `string` |  | ✓ |  | the cron job id |
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| cronJob | `body` | [Cron](#cron) | `models.Cron` | | ✓ | | the cron job data |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-repos-owner-name-cron-cron-200) | OK | OK |  | [schema](#patch-repos-owner-name-cron-cron-200-schema) |

#### Responses


##### <span id="patch-repos-owner-name-cron-cron-200"></span> 200 - OK
Status: OK

###### <span id="patch-repos-owner-name-cron-cron-200-schema"></span> Schema
   
  

[Cron](#cron)

### <span id="patch-repos-owner-name-registry-registry"></span> Update a named registry (*PatchReposOwnerNameRegistryRegistry*)

```
PATCH /api/repos/{owner}/{name}/registry/{registry}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| registry | `path` | string | `string` |  | ✓ |  | the registry name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| registryData | `body` | [Registry](#registry) | `models.Registry` | | ✓ | | the attributes for the registry |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-repos-owner-name-registry-registry-200) | OK | OK |  | [schema](#patch-repos-owner-name-registry-registry-200-schema) |

#### Responses


##### <span id="patch-repos-owner-name-registry-registry-200"></span> 200 - OK
Status: OK

###### <span id="patch-repos-owner-name-registry-registry-200-schema"></span> Schema
   
  

[Registry](#registry)

### <span id="patch-repos-owner-name-secrets-secret-name"></span> Update a named secret (*PatchReposOwnerNameSecretsSecretName*)

```
PATCH /api/repos/{owner}/{name}/secrets/{secretName}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| secretName | `path` | string | `string` |  | ✓ |  | the secret name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| secret | `body` | [Secret](#secret) | `models.Secret` | | ✓ | | the secret itself |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-repos-owner-name-secrets-secret-name-200) | OK | OK |  | [schema](#patch-repos-owner-name-secrets-secret-name-200-schema) |

#### Responses


##### <span id="patch-repos-owner-name-secrets-secret-name-200"></span> 200 - OK
Status: OK

###### <span id="patch-repos-owner-name-secrets-secret-name-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="patch-secrets-secret"></span> Update a global secret by name (*PatchSecretsSecret*)

```
PATCH /api/secrets/{secret}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| secret | `path` | string | `string` |  | ✓ |  | the secret's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| secretData | `body` | [Secret](#secret) | `models.Secret` | | ✓ | | the secret's data |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-secrets-secret-200) | OK | OK |  | [schema](#patch-secrets-secret-200-schema) |

#### Responses


##### <span id="patch-secrets-secret-200"></span> 200 - OK
Status: OK

###### <span id="patch-secrets-secret-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="patch-users-login"></span> Change a user (*PatchUsersLogin*)

```
PATCH /api/users/{login}
```

Changes the data of an existing user. Requires admin rights.

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| login | `path` | string | `string` |  | ✓ |  | the user's login name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| user | `body` | [User](#user) | `models.User` | | ✓ | | the user's data |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#patch-users-login-200) | OK | OK |  | [schema](#patch-users-login-200-schema) |

#### Responses


##### <span id="patch-users-login-200"></span> 200 - OK
Status: OK

###### <span id="patch-users-login-200-schema"></span> Schema
   
  

[User](#user)

### <span id="post-agents"></span> Create a new agent with a random token so a new agent can connect to the server (*PostAgents*)

```
POST /api/agents
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| agent | `body` | [Agent](#agent) | `models.Agent` | | ✓ | | the agent's data (only 'name' and 'no_schedule' are read) |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-agents-200) | OK | OK |  | [schema](#post-agents-200-schema) |

#### Responses


##### <span id="post-agents-200"></span> 200 - OK
Status: OK

###### <span id="post-agents-200-schema"></span> Schema
   
  

[Agent](#agent)

### <span id="post-debug-pprof-symbol"></span> Get pprof program counters mapping to function names (*PostDebugPprofSymbol*)

```
POST /api/debug/pprof/symbol
```

Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
Looks up the program counters listed in the request,
responding with a table mapping program counters to function names.
The requested program counters can be provided via GET + query parameters,
or POST + body parameters. Program counters shall be space delimited.

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-debug-pprof-symbol-200) | OK | OK |  | [schema](#post-debug-pprof-symbol-200-schema) |

#### Responses


##### <span id="post-debug-pprof-symbol-200"></span> 200 - OK
Status: OK

###### <span id="post-debug-pprof-symbol-200-schema"></span> Schema

### <span id="post-hook"></span> Incoming webhook from Github or Gitea (*PostHook*)

```
POST /api/hook
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| hook | `body` | [interface{}](#interface) | `interface{}` | | ✓ | | the webhook payload; Github or Gitea is automatically detected |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-hook-200) | OK | OK |  | [schema](#post-hook-200-schema) |

#### Responses


##### <span id="post-hook-200"></span> 200 - OK
Status: OK

###### <span id="post-hook-200-schema"></span> Schema

### <span id="post-log-level"></span> Set log level (*PostLogLevel*)

```
POST /api/log-level
```

Endpoint sets the current logging level. Requires admin rights.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| log-level | `body` | [PostLogLevelBody](#post-log-level-body) | `PostLogLevelBody` | | ✓ | | the new log level, one of <debug,trace,info,warn,error,fatal,panic,disabled> |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-log-level-200) | OK | OK |  | [schema](#post-log-level-200-schema) |

#### Responses


##### <span id="post-log-level-200"></span> 200 - OK
Status: OK

###### <span id="post-log-level-200-schema"></span> Schema
   
  

[PostLogLevelOKBody](#post-log-level-o-k-body)

###### Inlined models

**<span id="post-log-level-body"></span> PostLogLevelBody**


  


* composed type [PostLogLevelParamsBodyAllOf0](#post-log-level-params-body-all-of0)
* inlined member (*PostLogLevelParamsBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| log-level | string| `string` |  | |  |  |



**<span id="post-log-level-o-k-body"></span> PostLogLevelOKBody**


  


* composed type [PostLogLevelOKBodyAllOf0](#post-log-level-o-k-body-all-of0)
* inlined member (*postLogLevelOKBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| log-level | string| `string` |  | |  |  |



**<span id="post-log-level-o-k-body-all-of0"></span> PostLogLevelOKBodyAllOf0**


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| PostLogLevelOKBodyAllOf0 | string| string | |  |  |



**<span id="post-log-level-params-body-all-of0"></span> PostLogLevelParamsBodyAllOf0**


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| PostLogLevelParamsBodyAllOf0 | string| string | |  |  |



### <span id="post-orgs-owner-secrets"></span> Persist/create an organization secret (*PostOrgsOwnerSecrets*)

```
POST /api/orgs/{owner}/secrets
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| owner | `path` | string | `string` |  | ✓ |  | the owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| secretData | `body` | [Secret](#secret) | `models.Secret` | | ✓ | | the new secret |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-orgs-owner-secrets-200) | OK | OK |  | [schema](#post-orgs-owner-secrets-200-schema) |

#### Responses


##### <span id="post-orgs-owner-secrets-200"></span> 200 - OK
Status: OK

###### <span id="post-orgs-owner-secrets-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="post-queue-pause"></span> Pause a pipeline queue (*PostQueuePause*)

```
POST /api/queue/pause
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-queue-pause-200) | OK | OK |  | [schema](#post-queue-pause-200-schema) |

#### Responses


##### <span id="post-queue-pause-200"></span> 200 - OK
Status: OK

###### <span id="post-queue-pause-200-schema"></span> Schema

### <span id="post-queue-resume"></span> Resume a pipeline queue (*PostQueueResume*)

```
POST /api/queue/resume
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-queue-resume-200) | OK | OK |  | [schema](#post-queue-resume-200-schema) |

#### Responses


##### <span id="post-queue-resume-200"></span> 200 - OK
Status: OK

###### <span id="post-queue-resume-200-schema"></span> Schema

### <span id="post-repos-owner-name"></span> Activate a repository (*PostReposOwnerName*)

```
POST /api/repos/{owner}/{name}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-200) | OK | OK |  | [schema](#post-repos-owner-name-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-200-schema"></span> Schema
   
  

[Repo](#repo)

### <span id="post-repos-owner-name-chown"></span> Change a repository's owner, to the one holding the access token (*PostReposOwnerNameChown*)

```
POST /api/repos/{owner}/{name}/chown
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-chown-200) | OK | OK |  | [schema](#post-repos-owner-name-chown-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-chown-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-chown-200-schema"></span> Schema
   
  

[Repo](#repo)

### <span id="post-repos-owner-name-cron"></span> Persist/creat a cron job (*PostReposOwnerNameCron*)

```
POST /api/repos/{owner}/{name}/cron
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| cronJob | `body` | [Cron](#cron) | `models.Cron` | | ✓ | | the new cron job |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-cron-200) | OK | OK |  | [schema](#post-repos-owner-name-cron-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-cron-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-cron-200-schema"></span> Schema
   
  

[Cron](#cron)

### <span id="post-repos-owner-name-cron-cron"></span> Start a cron job now (*PostReposOwnerNameCronCron*)

```
POST /api/repos/{owner}/{name}/cron/{cron}
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| cron | `path` | string | `string` |  | ✓ |  | the cron job id |
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-cron-cron-200) | OK | OK |  | [schema](#post-repos-owner-name-cron-cron-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-cron-cron-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-cron-cron-200-schema"></span> Schema
   
  

[Pipeline](#pipeline)

### <span id="post-repos-owner-name-logs-number"></span> Deletes log (*PostReposOwnerNameLogsNumber*)

```
POST /api/repos/{owner}/{name}/logs/{number}
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-logs-number-200) | OK | OK |  | [schema](#post-repos-owner-name-logs-number-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-logs-number-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-logs-number-200-schema"></span> Schema

### <span id="post-repos-owner-name-move"></span> Move a repository to a new owner (*PostReposOwnerNameMove*)

```
POST /api/repos/{owner}/{name}/move
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| to | `query` | string | `string` |  | ✓ |  | the username to move the repository to |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-move-200) | OK | OK |  | [schema](#post-repos-owner-name-move-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-move-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-move-200-schema"></span> Schema

### <span id="post-repos-owner-name-pipelines"></span> Run/trigger a pipelines (*PostReposOwnerNamePipelines*)

```
POST /api/repos/{owner}/{name}/pipelines
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| options | `body` | [PipelineOptions](#pipeline-options) | `models.PipelineOptions` | | ✓ | | the options for the pipeline to run |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-pipelines-200) | OK | OK |  | [schema](#post-repos-owner-name-pipelines-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-pipelines-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-pipelines-200-schema"></span> Schema
   
  

[Pipeline](#pipeline)

### <span id="post-repos-owner-name-pipelines-number"></span> Restart a pipeline (*PostReposOwnerNamePipelinesNumber*)

```
POST /api/repos/{owner}/{name}/pipelines/{number}
```

Restarts a pipeline optional with altered event, deploy or environment

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| deploy_to | `query` | string | `string` |  |  |  | override the target deploy value |
| event | `query` | string | `string` |  |  |  | override the event type |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-pipelines-number-200) | OK | OK |  | [schema](#post-repos-owner-name-pipelines-number-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-pipelines-number-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-pipelines-number-200-schema"></span> Schema
   
  

[Pipeline](#pipeline)

### <span id="post-repos-owner-name-pipelines-number-approve"></span> Start pipelines in gated repos (*PostReposOwnerNamePipelinesNumberApprove*)

```
POST /api/repos/{owner}/{name}/pipelines/{number}/approve
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-pipelines-number-approve-200) | OK | OK |  | [schema](#post-repos-owner-name-pipelines-number-approve-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-pipelines-number-approve-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-pipelines-number-approve-200-schema"></span> Schema
   
  

[Pipeline](#pipeline)

### <span id="post-repos-owner-name-pipelines-number-cancel"></span> Cancels a pipeline (*PostReposOwnerNamePipelinesNumberCancel*)

```
POST /api/repos/{owner}/{name}/pipelines/{number}/cancel
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-pipelines-number-cancel-200) | OK | OK |  | [schema](#post-repos-owner-name-pipelines-number-cancel-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-pipelines-number-cancel-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-pipelines-number-cancel-200-schema"></span> Schema

### <span id="post-repos-owner-name-pipelines-number-decline"></span> Decline pipelines in gated repos (*PostReposOwnerNamePipelinesNumberDecline*)

```
POST /api/repos/{owner}/{name}/pipelines/{number}/decline
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| number | `path` | integer | `int64` |  | ✓ |  | the number of the pipeline |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-pipelines-number-decline-200) | OK | OK |  | [schema](#post-repos-owner-name-pipelines-number-decline-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-pipelines-number-decline-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-pipelines-number-decline-200-schema"></span> Schema
   
  

[Pipeline](#pipeline)

### <span id="post-repos-owner-name-registry"></span> Persist/create a registry (*PostReposOwnerNameRegistry*)

```
POST /api/repos/{owner}/{name}/registry
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| registry | `body` | [Registry](#registry) | `models.Registry` | | ✓ | | the new registry data |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-registry-200) | OK | OK |  | [schema](#post-repos-owner-name-registry-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-registry-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-registry-200-schema"></span> Schema
   
  

[Registry](#registry)

### <span id="post-repos-owner-name-repair"></span> Repair a repository (*PostReposOwnerNameRepair*)

```
POST /api/repos/{owner}/{name}/repair
```

#### Produces
  * text/plain

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-repair-200) | OK | OK |  | [schema](#post-repos-owner-name-repair-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-repair-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-repair-200-schema"></span> Schema

### <span id="post-repos-owner-name-secrets"></span> Persist/create a secret (*PostReposOwnerNameSecrets*)

```
POST /api/repos/{owner}/{name}/secrets
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| name | `path` | string | `string` |  | ✓ |  | the repository name |
| owner | `path` | string | `string` |  | ✓ |  | the repository owner's name |
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| secret | `body` | [Secret](#secret) | `models.Secret` | | ✓ | | the new secret |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-repos-owner-name-secrets-200) | OK | OK |  | [schema](#post-repos-owner-name-secrets-200-schema) |

#### Responses


##### <span id="post-repos-owner-name-secrets-200"></span> 200 - OK
Status: OK

###### <span id="post-repos-owner-name-secrets-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="post-secrets"></span> Persist/create a global secret (*PostSecrets*)

```
POST /api/secrets
```

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| secret | `body` | [Secret](#secret) | `models.Secret` | | ✓ | | the secret object data |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-secrets-200) | OK | OK |  | [schema](#post-secrets-200-schema) |

#### Responses


##### <span id="post-secrets-200"></span> 200 - OK
Status: OK

###### <span id="post-secrets-200-schema"></span> Schema
   
  

[Secret](#secret)

### <span id="post-user-token"></span> tbd (*PostUserToken*)

```
POST /api/user/token
```

tbd.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-user-token-200) | OK | OK |  | [schema](#post-user-token-200-schema) |

#### Responses


##### <span id="post-user-token-200"></span> 200 - OK
Status: OK

###### <span id="post-user-token-200-schema"></span> Schema

### <span id="post-users"></span> Create a user (*PostUsers*)

```
POST /api/users
```

Creates a new user account with the specified external login. Requires admin rights.

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| Authorization | `header` | string | `string` |  | ✓ | `"Bearer \u003cpersonal access token\u003e"` | Insert your personal access token |
| user | `body` | [User](#user) | `models.User` | | ✓ | | the user's data |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-users-200) | OK | OK |  | [schema](#post-users-200-schema) |

#### Responses


##### <span id="post-users-200"></span> 200 - OK
Status: OK

###### <span id="post-users-200-schema"></span> Schema
   
  

[User](#user)

## Models

### <span id="agent"></span> Agent


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| backend | string| `string` |  | |  |  |
| capacity | integer| `int64` |  | |  |  |
| created | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| last_contact | integer| `int64` |  | |  |  |
| name | string| `string` |  | |  |  |
| no_schedule | boolean| `bool` |  | |  |  |
| owner_id | integer| `int64` |  | |  |  |
| platform | string| `string` |  | |  |  |
| token | string| `string` |  | |  |  |
| updated | integer| `int64` |  | |  |  |
| version | string| `string` |  | |  |  |



### <span id="config"></span> Config


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | []integer| `[]int64` |  | |  |  |
| hash | string| `string` |  | |  |  |
| name | string| `string` |  | |  |  |



### <span id="cron"></span> Cron


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| branch | string| `string` |  | |  |  |
| created_at | integer| `int64` |  | |  |  |
| creator_id | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| name | string| `string` |  | |  |  |
| next_exec | integer| `int64` |  | |  |  |
| repo_id | integer| `int64` |  | |  |  |
| schedule | string| `string` |  | | @weekly,	3min, ... |  |



### <span id="feed"></span> Feed


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| author | string| `string` |  | |  |  |
| author_avatar | string| `string` |  | |  |  |
| author_email | string| `string` |  | |  |  |
| branch | string| `string` |  | |  |  |
| commit | string| `string` |  | |  |  |
| created_at | integer| `int64` |  | |  |  |
| event | string| `string` |  | |  |  |
| finished_at | integer| `int64` |  | |  |  |
| full_name | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| message | string| `string` |  | |  |  |
| name | string| `string` |  | |  |  |
| number | integer| `int64` |  | |  |  |
| owner | string| `string` |  | |  |  |
| ref | string| `string` |  | |  |  |
| refspec | string| `string` |  | |  |  |
| remote | string| `string` |  | |  |  |
| started_at | integer| `int64` |  | |  |  |
| status | string| `string` |  | |  |  |
| title | string| `string` |  | |  |  |



### <span id="file"></span> File


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| failed | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| mime | string| `string` |  | |  |  |
| name | string| `string` |  | |  |  |
| passed | integer| `int64` |  | |  |  |
| pid | integer| `int64` |  | |  |  |
| size | integer| `int64` |  | |  |  |
| skipped | integer| `int64` |  | |  |  |
| step_id | integer| `int64` |  | |  |  |
| time | integer| `int64` |  | |  |  |



### <span id="org-perm"></span> OrgPerm


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| admin | boolean| `bool` |  | |  |  |
| member | boolean| `bool` |  | |  |  |



### <span id="perm"></span> Perm


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| admin | boolean| `bool` |  | |  |  |
| created | integer| `int64` |  | |  |  |
| pull | boolean| `bool` |  | |  |  |
| push | boolean| `bool` |  | |  |  |
| synced | integer| `int64` |  | |  |  |
| updated | integer| `int64` |  | |  |  |



### <span id="pipeline"></span> Pipeline


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| author | string| `string` |  | |  |  |
| author_avatar | string| `string` |  | |  |  |
| author_email | string| `string` |  | |  |  |
| branch | string| `string` |  | |  |  |
| changed_files | []string| `[]string` |  | |  |  |
| clone_url | string| `string` |  | |  |  |
| commit | string| `string` |  | |  |  |
| created_at | integer| `int64` |  | |  |  |
| deploy_to | string| `string` |  | |  |  |
| enqueued_at | integer| `int64` |  | |  |  |
| error | string| `string` |  | |  |  |
| event | [WebhookEvent](#webhook-event)| `WebhookEvent` |  | |  |  |
| files | [][File](#file)| `[]*File` |  | |  |  |
| finished_at | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| link_url | string| `string` |  | |  |  |
| message | string| `string` |  | |  |  |
| number | integer| `int64` |  | |  |  |
| parent | integer| `int64` |  | |  |  |
| pr_labels | []string| `[]string` |  | |  |  |
| ref | string| `string` |  | |  |  |
| refspec | string| `string` |  | |  |  |
| reviewed_at | integer| `int64` |  | |  |  |
| reviewed_by | string| `string` |  | |  |  |
| sender | string| `string` |  | | uses reported user for webhooks and name of cron for cron pipelines |  |
| signed | boolean| `bool` |  | | deprecate |  |
| started_at | integer| `int64` |  | |  |  |
| status | [StatusValue](#status-value)| `StatusValue` |  | |  |  |
| steps | [][Step](#step)| `[]*Step` |  | |  |  |
| timestamp | integer| `int64` |  | |  |  |
| title | string| `string` |  | |  |  |
| updated_at | integer| `int64` |  | |  |  |
| variables | map of string| `map[string]string` |  | |  |  |
| verified | boolean| `bool` |  | | deprecate |  |



### <span id="pipeline-options"></span> PipelineOptions


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| branch | string| `string` |  | |  |  |
| variables | map of string| `map[string]string` |  | |  |  |



### <span id="pull-request"></span> PullRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| index | integer| `int64` |  | |  |  |
| title | string| `string` |  | |  |  |



### <span id="registry"></span> Registry


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| address | string| `string` |  | |  |  |
| email | string| `string` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| password | string| `string` |  | |  |  |
| token | string| `string` |  | |  |  |
| username | string| `string` |  | |  |  |



### <span id="repo"></span> Repo


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| active | boolean| `bool` |  | |  |  |
| allow_pr | boolean| `bool` |  | |  |  |
| avatar_url | string| `string` |  | |  |  |
| cancel_previous_pipeline_events | [][WebhookEvent](#webhook-event)| `[]WebhookEvent` |  | |  |  |
| clone_url | string| `string` |  | |  |  |
| config_file | string| `string` |  | |  |  |
| default_branch | string| `string` |  | |  |  |
| full_name | string| `string` |  | |  |  |
| gated | boolean| `bool` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| link_url | string| `string` |  | |  |  |
| name | string| `string` |  | |  |  |
| netrc_only_trusted | boolean| `bool` |  | |  |  |
| owner | string| `string` |  | |  |  |
| private | boolean| `bool` |  | |  |  |
| scm | [SCMKind](#s-c-m-kind)| `SCMKind` |  | |  |  |
| timeout | integer| `int64` |  | |  |  |
| trusted | boolean| `bool` |  | |  |  |
| visibility | [RepoVisibility](#repo-visibility)| `RepoVisibility` |  | |  |  |



### <span id="repo-patch"></span> RepoPatch


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| allow_pr | boolean| `bool` |  | |  |  |
| cancel_previous_pipeline_events | [][WebhookEvent](#webhook-event)| `[]WebhookEvent` |  | |  |  |
| config_file | string| `string` |  | |  |  |
| gated | boolean| `bool` |  | |  |  |
| netrc_only_trusted | boolean| `bool` |  | |  |  |
| timeout | integer| `int64` |  | |  |  |
| trusted | boolean| `bool` |  | |  |  |
| visibility | string| `string` |  | |  |  |



### <span id="repo-visibility"></span> RepoVisibility


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| RepoVisibility | string| string | |  |  |



### <span id="s-c-m-kind"></span> SCMKind


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| SCMKind | string| string | |  |  |



### <span id="secret"></span> Secret


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| event | [][WebhookEvent](#webhook-event)| `[]WebhookEvent` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| image | []string| `[]string` |  | |  |  |
| name | string| `string` |  | |  |  |
| plugins_only | boolean| `bool` |  | |  |  |
| value | string| `string` |  | |  |  |



### <span id="status-value"></span> StatusValue


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| StatusValue | string| string | |  |  |



### <span id="step"></span> Step


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| agent_id | integer| `int64` |  | |  |  |
| children | [][Step](#step)| `[]*Step` |  | |  |  |
| end_time | integer| `int64` |  | |  |  |
| environ | map of string| `map[string]string` |  | |  |  |
| error | string| `string` |  | |  |  |
| exit_code | integer| `int64` |  | |  |  |
| id | integer| `int64` |  | |  |  |
| name | string| `string` |  | |  |  |
| pgid | integer| `int64` |  | |  |  |
| pid | integer| `int64` |  | |  |  |
| pipeline_id | integer| `int64` |  | |  |  |
| platform | string| `string` |  | |  |  |
| ppid | integer| `int64` |  | |  |  |
| start_time | integer| `int64` |  | |  |  |
| state | [StatusValue](#status-value)| `StatusValue` |  | |  |  |



### <span id="task"></span> Task


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| agent_id | integer| `int64` |  | |  |  |
| data | []integer| `[]int64` |  | |  |  |
| dep_status | map of [StatusValue](#status-value)| `map[string]StatusValue` |  | |  |  |
| dependencies | []string| `[]string` |  | |  |  |
| id | string| `string` |  | |  |  |
| labels | map of string| `map[string]string` |  | |  |  |
| run_on | []string| `[]string` |  | |  |  |



### <span id="user"></span> User


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| admin | boolean| `bool` |  | | Admin indicates the user is a system administrator.

NOTE: If the username is part of the WOODPECKER_ADMIN
environment variable this value will be set to true on login. |  |
| avatar_url | string| `string` |  | | the avatar url for this user. |  |
| email | string| `string` |  | | Email is the email address for this user.

required: true |  |
| id | integer| `int64` |  | | the id for this user.

required: true |  |
| login | string| `string` |  | | Login is the username for this user.

required: true |  |



### <span id="webhook-event"></span> WebhookEvent


  

| Name | Type | Go type | Default | Description | Example |
|------|------|---------| ------- |-------------|---------|
| WebhookEvent | string| string | |  |  |


