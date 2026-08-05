# Forges

## Supported features

| Feature                                                                                                                | [GitHub](20-github.md) | [Gitea](30-gitea.md) | [Forgejo](35-forgejo.md) | [Gitlab](40-gitlab.md) | [Bitbucket](50-bitbucket.md) | [Bitbucket Datacenter](60-bitbucket_datacenter.md) |
| ---------------------------------------------------------------------------------------------------------------------- | ---------------------- | -------------------- | ------------------------ | ---------------------- | ---------------------------- | -------------------------------------------------- |
| Event: Push                                                                                                            | :white_check_mark:     | :white_check_mark:   | :white_check_mark:       | :white_check_mark:     | :white_check_mark:           | :white_check_mark:                                 |
| Event: Tag                                                                                                             | :white_check_mark:     | :white_check_mark:   | :white_check_mark:       | :white_check_mark:     | :white_check_mark:           | :white_check_mark:                                 |
| Event: Pull-Request                                                                                                    | :white_check_mark:     | :white_check_mark:   | :white_check_mark:       | :white_check_mark:     | :white_check_mark:           | :white_check_mark:                                 |
| Event: Release                                                                                                         | :white_check_mark:     | :white_check_mark:   | :white_check_mark:       | :white_check_mark:     | :x:                          | :x:                                                |
| Event: Deploy¹                                                                                                         | :white_check_mark:     | :x:                  | :x:                      | :x:                    | :x:                          | :x:                                                |
| [Event: Pull-Request-Metadata](../../../20-usage/50-environment.md#pull_request_metadata-specific-event-reason-values) | :white_check_mark:     | :white_check_mark:   | :white_check_mark:       | :white_check_mark:     | :x:                          | :x:                                                |
| [Multiple workflows](../../../20-usage/25-workflows.md)                                                                | :white_check_mark:     | :white_check_mark:   | :white_check_mark:       | :white_check_mark:     | :white_check_mark:           | :white_check_mark:                                 |
| [when.path filter](../../../20-usage/20-workflow-syntax.md#path)                                                       | :white_check_mark:     | :white_check_mark:   | :white_check_mark:       | :white_check_mark:     | :white_check_mark:           | :white_check_mark:                                 |

¹ The deployment event can be triggered for all forges from Woodpecker directly. However, only GitHub can trigger them using webhooks.

In addition to this, Woodpecker supports [addon forges](../100-addons.md) if the forge you are using does not meet the [Woodpecker requirements](../../../92-development/02-core-ideas.md#forges) or your setup is too specific to be included in the Woodpecker core.

## Multiple forges

:::danger
Support for connecting multiple forges is not finished yet and has to be considered experimental. Please do not use it in public or otherwise untrusted environments. See [known limitations](#known-limitations) below.
:::

Only **one** forge can be configured using environment variables, for example [GitHub](20-github.md) or [GitLab](40-gitlab.md). Enabling several forge drivers at once does not create several forges. All other forges have to be added by an admin under `Settings` -> `Forges` and exist in the database only. The forge from the environment is written back on every server start, so changing its driver replaces it instead of adding a new one.

### Restricting who can log in

[`WOODPECKER_ORGS`](../10-server.md#orgs) applies to all connected forges. As organization names are only unique within a single forge, someone could create an organization with an approved name on another connected forge to gain access.
Each forge can therefore carry its own list under `Settings` -> `Forges` -> `Advanced options` -> `Allowed organizations`, which replaces the global setting for logins using that forge. Forges without an own list keep using `WOODPECKER_ORGS`; if neither is set, organizations are not checked at all.

Use the organization name, for GitLab the group's [`full_path`](https://docs.gitlab.com/api/groups/) like `my-group/my-subgroup`. Matching is case-insensitive and only happens while logging in. Note that a forge can hide organizations from Woodpecker although the user is a member of them, see GitHub and [GitLab](40-gitlab.md#allowed-organizations).

### Known limitations

- [`WOODPECKER_ADMIN`](../10-server.md#admin) is matched by login name only. A user with a matching login on **any** connected forge becomes an admin.
- Users are stored per forge, so the same person logging in through two forges gets two independent accounts.
