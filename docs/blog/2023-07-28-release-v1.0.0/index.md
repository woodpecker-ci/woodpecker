---
title: Presenting Woodpecker 1.0.0
description: Introducing Woodpecker 1.0.0 and its new features.
slug: release-v1.0.0
authors:
  - name: 6543
    title: Maintainer of Woodpecker
    url: https://github.com/6543
    image_url: https://github.com/6543.png
tags: [release, stable]
hide_table_of_contents: false
---

We are proud to present you Woodpecker v1.0.0.
It took us quite some time, but now we are sure it's ready, and so you should really have a look at it.

<!--truncate-->

We did refactor a lot of code, so maintaining the codebase should be much easier.
There were also added a ton of bug fixes and enhancements, not to mention some long-awaited features.
With this, you should be able to significantly improve and streamline your code pipelines,
empowering you to automate and optimize workflows like never before.

## Some picked highlights:

### Add Support for Cron Jobs

Automate recurring tasks with ease using Woodpecker's cron jobs feature.
Schedule pipelines to run at specified intervals or times, optimizing repetitive workflows.
[Read more](/docs/usage/cron)

### YAML Map Merge, Overrides, and Sequence Merge Support

With enhanced YAML support, managing complex configurations becomes a breeze. Merge maps, apply overrides, and sequence merging—all within your YAML files.
This is providing greater flexibility and control over your pipelines.
[Read more](/docs/usage/advanced-yaml-syntax)

### Add Web-UI for Admins

The new Admin UI simplifies administration tasks, making it easier for admins to manage user accounts, agents and tasks.
Now, you can effortlessly add new agents or pause the task queue to perform maintenance.
![Image of admin queue view](admin_queue_ui.png)

### Localize Web-UI

Woodpecker embraces global diversity by allowing users to change their locale in the user settings.
Enjoy a personalized experience with the language of your choice when interacting.
If your language is not available or only partial translated, have a look at our [Weblate](https://translate.woodpecker-ci.org/engage/woodpecker-ci/).

### Add Evaluate to When Filter

Enhance pipeline flexibility with the new when evaluate filter, enabling or disable steps conditional based on custom conditions.
Tailor your workflows to respond dynamically to specific triggers and events.
[Read more](/docs/usage/pipeline-syntax#evaluate)

### Global and Organization Secrets

Save time and effort by declaring secrets for a whole instance or organization.
Simplify your workflow and securely manage sensitive information across projects.

### Pipeline Log Output Download

Retrieve pipeline log outputs effortlessly, empowering you to analyze and troubleshoot your pipelines.

## Changelog

The whole changelog can be viewed in our project source folder at [CHANGELOG.md](https://github.com/woodpecker-ci/woodpecker/blob/v1.0.0/CHANGELOG.md)
