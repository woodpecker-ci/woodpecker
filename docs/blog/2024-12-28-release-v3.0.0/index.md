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

> Disclaimer upfront: The Woodpecker team is aware that this release contains _a lot_ of changes, also many which force users to update their pipeline definitions.
> We understand that this can be a tedious task, especially when managing numerous repositories and pipelines. Each of these changes was carefully considered and thoroughly discussed beforehand, with specific reasoning behind every decision.
> A significant portion of these updates is focused on “breaking free” from outdated and suboptimal Drone definitions.
> Thank you for your patience and understanding as we implement these essential breaking changes!

Security has been the major focus in this major release.
Several known vulnerabilities were patched (and also backported to v2 releases), ensuring that your CI/CD environment is protected from potential exploits.
The secrets handling mechanism has also been upgraded, preventing accidental leaks and making it simpler to keep sensitive information fully encrypted.
Specifically, `secrets:` have been deprecated in favor of central syntax to specify secrets: `from_secret`.
This new way provides more flexibility (by being to use different names for the source and destination secrets) and ensure a safe internal secret parsing through a unified engine.
Because `secrets:` were nothing else than an env var in the end, it removes potential confusion about the differences between values specified in `environment:` and `secrets`.
Beforehand, users specified both `secrets:` and `environment:` and these were then merged behind the scenes.
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

## Support for rootless images

FIXME

## UI

We squashed many UI related bug fixes in this release.
Many were small misalignment related to padding, margins or other edge cases for smaller screen sizes.
We also aimed to harmonize the icons across the UI, specifically across logical subgroups, such as status-icons or admin panel icons.
Last, UI elements are now sized in a relative way, meaning they will all scale relative when you change the font-size or zoom in/out.

## Enhanced debugging: rerun failed workflows locally

FIXME

## Enhanced granular control over Pull Request approvals

## Delete old pipeline logs in DB through the CLI

FIXME

## Migration to 5-char CRON syntax

## Notable bug fixes

:::info
All fixes highlighted here have been backported to v2.x.
:::

GitLab support has been improved by a lot, i.e., many bugs have been fixed which caused WP to be practically unusable with subgroups.

A panic caused by an unreachable forge has been patched.
Yet, the occasional error of "pipeline definition not found" is not yet fully understood or fixed.
We are aware of it and have been discussing it among maintainers extensively.
