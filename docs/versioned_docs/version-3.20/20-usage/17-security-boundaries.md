# Security boundaries

Woodpecker runs code from your repositories. This page lists who can do what, with which privileges, and which secrets and credentials a pipeline or step gets.

## What Woodpecker trusts the forge with

Woodpecker does not check this information itself, it uses what the [forge](../30-administration/10-configuration/12-forges/11-overview.md) reports:

- Who a user is, their username and their organizations.
- Which permissions a user has on a repository (read, push or admin).
  A user who can sign in to Woodpecker gets the same [permissions](#permissions) there.
- The webhook data, like the event, the author and if a pull request comes from a fork.
- That only users with write access can push branches to the repository itself.

Woodpecker assumes that the information from the forge is correct. There is no further validation.

<!-- TODO: document forge specific behavior, e.g. which login is used as pipeline author on each forge (GitLab releases use the name of the commit author). -->

### Webhooks

Woodpecker generates a random secret for every repository when it is activated.
The webhook URL carries a non-expiring token signed with this secret, which names the forge and repository and is checked on each received hook.
A secret only allows tokens for its own repository.

The token does not cover the body, so keep the webhook URL private: whoever has it can send forged requests for its repository.
<!-- TODO: Gitea, Forgejo and Bitbucket Data Center also sign the body with the secret and Woodpecker verifies it, so do not leak the secret either. (only Bitbucket Data Center is verified for now) -->

This only protects against a leaked URL, not against admins of the repository on the forge, who can read the URL and, depending on your forge, the secret.

:::info
If the webhook URL or secret got leaked, use **Repair repository** in the repository settings to rotate the secret.
:::

A valid token only proves which repository a webhook is for.
What the pipeline can do then depends on the author the forge reports, the [approval](#approval-of-pipelines) rules and the [credentials](#git-credentials) it gets.

## Permissions

- Instance admins ([`WOODPECKER_ADMIN`](../30-administration/10-configuration/10-server.md#admin)) are admins of every org and repository.
  Users in the list become instance admins when they sign in, the list is read at server start.
  Instance admins can also be granted and revoked in the admin settings or with the users API.
  Revoking a listed user in the UI takes effect right away, but if you don't remove them from the list, they are admin again at their first sign in after the next server restart.
  With [multiple forges](../30-administration/10-configuration/12-forges/11-overview.md#known-limitations), a user with a listed login on any connected forge becomes admin.
- Admins of a repository on the forge can change its [project settings](./75-project-settings.md).
- Users with push access can start, restart, cancel, approve and decline pipelines, and manage crons.
- Users with read access can see pipelines and their logs. The [project visibility](./75-project-settings.md#project-visibility) can give read access to everybody.

Woodpecker caches the permissions of a user for up to [one hour](https://github.com/woodpecker-ci/woodpecker/blob/5d41af5c30bc11f52bbaee298d387191e16192c8/server/router/middleware/session/repo.go#L35).
If fetching permissions failed, the user is downgraded to read access till fetching works again.

## Approval of pipelines

A pipeline can wait for approval before it runs, see [require approval setting](./75-project-settings.md#require-approval-for).
By default this is the case for pull requests from forks.

- Pipelines of users in the allowed users list never need approval.
  The list is compared with the author login the forge reports.
- On most forges the author of a pull request is the user who opened it, so everybody who can push to it runs pipelines without approval.
- Manual pipelines, crons, restarts (also of declined pipelines) and deployments started in Woodpecker never need approval,
  as starting them needs push access like approving does.
  Their author is the user who started them, except for crons, which have none, and restarts, which keep the author of the original pipeline.
  <!-- TODO: restarts should get the user who restarted them as author, like deployments do since #7358. -->
- An approved pipeline runs the configuration it was created with, but gets the current secrets and the Git credentials of the user who approved it.
  <!-- https://github.com/woodpecker-ci/woodpecker/blob/1182744d09b1c097b47749ed813cc7e0549c25c9/server/pipeline/approve.go#L60 -->

## Workflow configuration

Woodpecker reads the workflow files from the commit a pipeline runs for, so a pull request runs the workflow of its own branch.
Everybody who can push or open a pull request that runs without approval can control the workflows.

Values from the forge, like the commit message, the author or the pull request labels, are untrusted input.
[String substitution](./50-environment.md#string-substitution) puts them into the YAML text before it is parsed.
Woodpecker quotes values with line breaks, so they cannot add steps or other YAML structure.
Where the commit data is not written by the one who wrote the workflow, treat it with care:

- A workflow that runs after a merge uses the commit message of the merge, which can be written by the contributor.
- A [config extension](./72-extensions/index.md) returns the workflow, but the commit data is still substituted into it.

## Secrets

Secrets are only passed to pipelines of the [events](./40-secrets.md#events-filter) they are enabled for.
Enabling `pull_request` gives them to everybody who can open a pull request that runs without approval. You can partially limit this by filtering for plugins.

A secret limited to [plugins](./40-secrets.md#plugins-filter) is only passed to those images, and only if the step is a [plugin](#plugins).
[In this case](./51-plugins/20-creating-plugins.md#security), the plugin is in charge of making sure it cannot be leaked.

Secrets from a [secret extension](./72-extensions/55-secret-extension.md) without events are passed to all events.

## Trusted repositories

Only instance admins can make a repository [trusted](./75-project-settings.md#trusted).
Each option allows more in the pipeline containers:

- Network: `dns`, `dns_search`, `extra_hosts` and `network_mode`.
- Volumes: `volumes`, `devices` and `tmpfs`.
- Security: `privileged` and the Git credentials in every step.

:::danger
Only enable this if you know and trust the ones who can run and alter pipeline configs,
as they can at least compromise the hosts agents run on!
:::

## Git credentials

For private repositories, or with [`WOODPECKER_AUTHENTICATE_PUBLIC_REPOS`](../30-administration/10-configuration/10-server.md#authenticate_public_repos),
Woodpecker creates Git credentials of a forge user. They can also push to the repository.

They are only passed to the default clone step, to plugins in the [trusted clone plugins](./75-project-settings.md#custom-trusted-clone-plugins)
and to every step of a repository with the [security option](#trusted-repositories).

## Plugins

Plugins can get more trust than other steps: plugin only secrets, [privileged plugins](../30-administration/10-configuration/10-server.md#plugins_privileged) and trusted clone plugins.
Everybody who can change the workflow controls their settings and the workspace.

A step is only a plugin if it sets none of `commands`, `entrypoint` and `environment`, otherwise it gets none of this trust.
The linter rejects `commands` and `entrypoint` next to `settings`, but only warns about `environment`.
<!-- TODO: the linter also rejects `environment` next to `settings` once https://github.com/woodpecker-ci/woodpecker/pull/7404 is merged (scheduled for the next major). -->

See [creating plugins](./51-plugins/20-creating-plugins.md#security) for what this means for plugin authors.

## Extensions

[Extensions](./72-extensions/index.md#security) get the repository and pipeline data, optionally the Git credentials,
and return configurations, secrets or registries. They are called before a pipeline is approved.

[Woodpecker signs its requests](./72-extensions/index.md#security) and only calls hosts allowed by [`WOODPECKER_EXTENSIONS_ALLOWED_HOSTS`](../30-administration/10-configuration/10-server.md#extensions_allowed_hosts).

## Agents and backends

Agents connect with a [system or agent token](../30-administration/10-configuration/30-agent.md#agent-registration),
using TLS only with [`WOODPECKER_GRPC_SECURE`](../30-administration/10-configuration/30-agent.md#grpc_secure).
The server serves gRPC without TLS, so put it behind a [TLS proxy](../30-administration/10-configuration/10-server.md#reverse-proxy) or use a [unix socket](../30-administration/10-configuration/10-server.md#grpc_addr) for agents on the same host.
Organization admins, and every user for their own repositories, can add own agents, unless [`WOODPECKER_DISABLE_USER_AGENT_REGISTRATION`](../30-administration/10-configuration/10-server.md#disable_user_agent_registration) is set.
Every admin of an organization can see the tokens of all its agents, also of the ones other admins added, and can edit and remove them.
This gives no more access than adding an own agent, as [organization agents](../30-administration/10-configuration/30-agent.md#agent-filters) only get workflows of their organization.
Instance-wide agents run workflows of all organizations, unless an organization allows [only its own agents](../30-administration/10-configuration/30-agent.md#organization-agents-only).

The backend isolates the steps from the agent. The [local backend](../30-administration/10-configuration/11-backends/30-local.md) has no isolation at all.
