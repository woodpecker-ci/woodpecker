---
title: It's time for some changes - Woodpecker 2.0.0
description: Introducing Woodpecker 2.0.0 with more than 350 changes
slug: release-v200
authors:
  - name: Anbraten
    title: Maintainer of Woodpecker
    url: https://github.com/anbraten
    image_url: https://github.com/anbraten.png
  - name: qwerty287
    title: Maintainer of Woodpecker
    url: https://github.com/qwerty287
    image_url: https://github.com/qwerty287.png
tags: [release, stable]
hide_table_of_contents: false
---

We are proud to present you Woodpecker v2.0.0 with more than 350 changes from our fabulous community. This release includes a lot of new features, improvements and some breaking changes which most of you probably already tested using the `next` tag or the RC version.

<!--truncate-->

It has been almost 4 months since the last major release and we collected quite some changes in the meantime. We decided to release a new major version instead of a minor one, because there are a few breaking changes and we want to make sure that everyone is aware of them.

## How we plan to handle releases in the future

In future, there won't be backports anymore as they require quite an amount of maintenance. Instead, we'll release our current state of the `main` branch with the correct version (according to semver) every few weeks. Of course, critical bug and security fixes are released as soon as possible. To not release new major versions too often, we'll try to hold back breaking changes for a longer time and release them all together in a new major version.

## Breaking changes

### Renamed some api routes

We renamed some API routes to be more consistent. So we suggest admins to update all repository webhooks by clicking on the newly added `Repair all repositories` button in the admin settings.

### Dropped deprecated environment variables and CLI commands

For v1.0.0, we deprecated a bunch of old environment variables like `CI_BUILD_*`. These variables were removed in this version, you therefore have to use the new ones.
Also, the deprecated `build` command of the CLI was removed. Use `pipeline` instead.

### Removed SSH backend

Due to various issues with the SSH backend we decided to remove it.
As an alternative, you can install an agent running the local backend directly on the remote machine or you can simply execute `ssh` commands connecting to the remote server in your pipeline.

### Deprecated `platform` filter

The `platform` filter has been removed. Use the more advanced labels instead ([read more](./docs/usage/workflow-syntax#filter-by-platform)).

### Update Docker to v24

We updated Docker to v24 as of some security patches. If you use an older version of Docker, you might need to upgrade it.

### Removed plugin-only option from secrets

Security is pretty important to us and we want to make sure that no one can steal your secrets. Therefore, we decided to remove the plugin-only option from secrets and instead, if you define an image filter, it will be automatically only available to plugins using the defined image names.

## Migration notes

There have been a few more breaking changes. [Read more about what you need to do when upgrading!](../docs/migrations#200)

## New features

But that's enough about breaking changes. Let's talk about the new features!

### Config errors and warnings in the UI

You ever wondered why a secret was not working and after hours of debugging you found out that you misspelled the secret name? Or you used a wrong key in your YAML config? Woodpecker now shows errors and linter warnings directly in it's UI, notifying you about missing secrets, incorrect configuration or deprecated settings!

![Image of warnings and errors in UI](./linter_warnings_errors.png)

### Repository and organization overview for admins

Admins now get an overview over all repositories and organizations registered on the server, allowing them to perform common actions like deleting directly from the admin dashboard.

![Image of repos overview](./admin_repos.png)

### Support for user secrets

It is now possible to add secrets for all repos owned by yourself, similar to organization and global secrets.

### Bitbucket cloud support for multi-workflows

We enhanced support for Bitbucket, allowing you to use multiple workflows just as you probably know from all other forges already.

### Full support for Kubernetes backend

Many of you already used it extensively in the past, but now we can finally call the Kubernetes backend ready for production use. Supporting all major features and even quite some Kubernetes specific options.

### Auto theme

The UI now supports automatically adapting the theme to your browser config, so no more light mode in the middle of the night!

### Update notification

Updates are awesome as they bring new features and bug fixes most of the time, but sometimes there are also important security fixes which should be installed as soon as possible. To not miss any of them, we added a notification to the UI for admins if there's a new update available.

## Changelog

The full changelog can be viewed in our project source folder at [CHANGELOG.md](https://github.com/woodpecker-ci/woodpecker/blob/v2.0.0/CHANGELOG.md)
