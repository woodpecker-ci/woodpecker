# Security boundaries

Woodpecker runs code from your repositories. This page lists who can do what and which privileges, secrets and credentials a pipeline gets. The linked pages have the details.

## What Woodpecker trusts the forge with

Woodpecker does not check this information itself, it uses what the [forge](../30-administration/10-configuration/12-forges/11-overview.md) reports:

- Who a user is, their login and their organizations.
- Which permissions a user has on a repository (read, push or admin).
- The webhook data, like the event, the author and if a pull request comes from a fork.
- That only users with write access can push branches to the repository itself.

If the forge reports something wrong, the boundaries on this page are wrong too.

<!-- TODO: document forge specific behavior, e.g. which login is used as pipeline author on each forge (GitLab releases use the name of the commit author). -->

## Permissions

- Instance admins ([`WOODPECKER_ADMIN`](../30-administration/10-configuration/10-server.md#admin)) are admins of every repository.
- Admins of a repository on the forge can change its [project settings](./75-project-settings.md).
- Users with push access can start, restart, cancel, approve and decline pipelines, and manage crons.
- Users with read access can see pipelines and their logs. The [project visibility](./75-project-settings.md#project-visibility) can give read access to everybody.

Woodpecker caches the permissions of a user for up to one hour.

## Approval of pipelines

A pipeline can wait for approval before it runs, see [require approval for](./75-project-settings.md#require-approval-for). By default this is the case for pull requests from forks.

- Pipelines of users in the allowed users list never need approval. The list is compared with the author login the forge reports.
- On most forges the author of a pull request is the user who opened it, so everybody who can push to it runs pipelines without approval.
- Manual pipelines, crons, restarts (also of declined pipelines) and deployments started in Woodpecker never need approval, as starting them needs push access like approving does.
- An approved pipeline runs the configuration it was created with, but gets the current secrets and the Git credentials of the user who approved it.

## Secrets

Secrets are only passed to pipelines of the [events](./40-secrets.md#events-filter) they are enabled for. Enabling `pull_request` gives them to everybody who can open a pull request that runs without approval.

A secret limited to [plugins](./40-secrets.md#plugins-filter) is only passed to these images. Secrets from a [secret extension](./72-extensions/55-secret-extension.md) without events are passed to all events.

## Trusted repositories

Only instance admins can make a repository [trusted](./75-project-settings.md#trusted). Each option allows more in the pipeline containers:

- Network: `dns`, `dns_search`, `extra_hosts` and `network_mode`.
- Volumes: `volumes`, `devices` and `tmpfs`.
- Security: `privileged` and the Git credentials in every step.

## Git credentials

For private repositories, or with [`WOODPECKER_AUTHENTICATE_PUBLIC_REPOS`](../30-administration/10-configuration/10-server.md#authenticate_public_repos), Woodpecker creates Git credentials of a forge user. They can also push to the repository.

They are only passed to the default clone step, to plugins in the [trusted clone plugins](./75-project-settings.md#custom-trusted-clone-plugins) and to every step of a repository with the security option.

## Plugins

Plugins can get more trust than other steps: plugin only secrets, [privileged plugins](../30-administration/10-configuration/10-server.md#plugins_privileged) and trusted clone plugins. Everybody who can change the workflow controls their settings and the workspace.

See [creating plugins](./51-plugins/20-creating-plugins.md#security) for what this means for plugin authors.

## Extensions

[Extensions](./72-extensions/index.md#security) get the repository and pipeline data, optionally the Git credentials, and return configurations, secrets or registries. They are called before a pipeline is approved.

Woodpecker signs its requests and only calls hosts allowed by [`WOODPECKER_EXTENSIONS_ALLOWED_HOSTS`](../30-administration/10-configuration/10-server.md#extensions_allowed_hosts).

## Agents and backends

Agents connect with a [system or agent token](../30-administration/10-configuration/30-agent.md#agent-registration), using TLS only with [`WOODPECKER_GRPC_SECURE`](../30-administration/10-configuration/30-agent.md#grpc_secure). Organization admins can add own agents, unless [`WOODPECKER_DISABLE_USER_AGENT_REGISTRATION`](../30-administration/10-configuration/10-server.md#disable_user_agent_registration) is set.

The backend isolates the steps from the agent. The [local backend](../30-administration/10-configuration/11-backends/30-local.md) has no isolation at all.
