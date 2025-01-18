---
title: Woodpecker 3.0.0
description: Introducing Woodpecker 2.5.0
slug: release-v300
authors:
  - name: pat-s
    title: Maintainer of Woodpecker
    url: https://github.com/pat-s
    image_url: https://github.com/pat-s.png
tags: [release, major]
hide_table_of_contents: false
---

:::
Disclaimer upfront: The Woodpecker team is aware that this release contains _a lot_ of changes, also many which force users to update their pipeline definitions.
We understand that this can be a tedious task, especially when managing numerous repositories and pipelines. Each change was carefully considered and thoroughly discussed, with specific reasoning behind every decision.
A significant portion of these updates is focused on “breaking free” from outdated and suboptimal Drone definitions.
Thank you for your patience and understanding as we implement these essential breaking changes!
:::

Security has been the major focus in this major release.
Besides patching known vulnerabilities (and also backporting these to v2 releases), the secrets handling mechanism has been improved, preventing accidental leaks and making it simpler to keep sensitive information fully encrypted.

Specifically, the `secrets:` keyword has been deprecated in favor of a more flexible (and secure) way to specify secrets: `from_secret:`.
This new approach provides more flexibility (by using different names for the source and destination secrets) and ensures a safe internal secret parsing through a unified engine.
Because secrets defined via `secrets:` were simple env vars in the end, this change also removes potential confusion about the differences between values specified in `environment:` and `secrets`.
Now, both are defined in `environment:` using an expressive syntax:

```yaml
steps:
  name:
    image: alpine
    commands:
      - echo "The secret is $TOKEN_ENV"
    environment:
      TOKEN_ENV:
        from_secret: SECRET_TOKEN
```

## Rootless images

Woodpecker now supports running rootless images by adjusting the entrypoints and directory permissions in the containers in a way that allows non-privileged users to execute tasks.

In addition, all images published by Woodpecker (Server, Agent, CLI) now use a non-privileged user (`woodpecker` with UID and GID 1000) by default.
## Register Your Own Agents for Users or Organizations [#3539](https://github.com/woodpecker-ci/woodpecker/pull/3539)
WoodpeckerCI now lets you register custom agents scoped to individual users or organizations. This means you can bring your own agents, configured to meet the unique needs of your projects, and assign them to specific users or organizational workflows.

This update provides flexibility for teams with diverse requirements, allowing them to integrate agents tailored to specific tasks or environments seamlessly into their pipelines.

## Replay Pipelines Locally Using `cli exec` [#4103](https://github.com/woodpecker-ci/woodpecker/pull/4103)
Debugging pipelines no longer requires endless small adjustments and repeated pushes. With the new `cli exec` feature, you can download pipeline metadata directly from the server and replay it locally. This allows you to test and fix issues in a replica of the server environment, all from your machine.

By enabling local debugging, this feature accelerates the development process and provides deeper insights into pipeline behavior without relying on server-side execution for every small change.
:::info
The agent image must remain rootful by default to be able to mount the Docker socket when Woodpecker is used with the `docker` backend.
The helm chart will start to use a non-privileged user by utilizing `securityContext`.
Running a completely rootless agent with the `docker` backend may be possible by using a rootless docker daemon.
However, this requires more work and is currently not supported.
:::

## UI

We have fixed many UI-related bugs in this version.
Many were small misalignment related to padding, margins or other edge cases related to small screen sizes.
We also aimed to harmonize the icons across the UI, specifically across logical subgroups, such as status-icons or admin panel icons.

UI elements are now sized in a relative way, meaning they will all scale relative when you change the font-size or zoom in/out.

## Enhanced debugging: rerun failed workflows locally

With Woodpecker 3.0 one has the option to rerun failed pipelines locally by starting these through the `woodpecker-cli` using the pipeline metadata.

![debug-pipelines-option](debug-pipelines.png)

:::info
In order to use this feature, all required pipeline elements must be passed, e.g. secrets.
However, secrets are not included in the pipeline metadata and must be passed manually to the local execution call.
:::

## Fine grained control over Pull Request approvals

New approval options for Pull Request workflows are available.
By default, Pull Requests pipelines from forks are not started automatically but require explicit approval.
This avoids potentially malicious PRs which could expose secrets or execute other unwanted tasks without the repo owner noticing it.

![screenshot of new approval-requirements options](approval-requirements.png)

## Deleting old pipeline logs

Deleting a pipeline now successfully also deletes its related logs.
Beforehand, there was an issue where the logs were not deleted and were kept in the DB forever.

You might want to check [#4572](https://github.com/woodpecker-ci/woodpecker/pull/4572) for more details including a snippet how to delete orphaned entries of a Postgres DB.

:::info
There is no option yet to auto-delete old pipeline logs after a specific time or event.
Please follow [#1068](https://github.com/woodpecker-ci/woodpecker/issues/1068) for future updates.
:::

## Migration to 5-char CRON syntax

The underlying CRON package was changed to one that now uses the (more common) 5-char CRON syntax.
Users needs to actively updated their CRON entries, otherwise existing pipelines will error during execution.

## Known Issues

The generic `pipeline definition not found` is still present and not yet understood.
This error message can be triggered by various elements (which the most likely one being a (temporary) connection issue with the forge) and the error return/output must be improved first in order to take appropriate action.
