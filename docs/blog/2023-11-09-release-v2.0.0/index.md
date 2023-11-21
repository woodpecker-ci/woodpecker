---
title: It's time for some changes - Woodpecker 2.0.0
description: Introducing Woodpecker 2.0.0 with > 300 changes
slug: release-v200
draft: true
authors:
  - name: Anbraten
    title: Maintainer of Woodpecker
    url: https://github.com/anbraten
    image_url: https://github.com/anbraten.png
tags: [release, stable]
hide_table_of_contents: false
---

We are proud to present you Woodpecker v2.0.0 with more than 300 changes from our great community. This release includes a lot of new features, improvements and some breaking changes which most of you probably already tested using the `next` tag.

<!--truncate-->

## Breaking changes

- Use int64 for IDs in woodpecker client lib [#2703]
- Woodpecker-go: Use Feed instead Activity [#2690]
- Do not sanitzie secrets with 3 or less chars [#2680]
- fix(deps): update docker to v24 [#2675]
- Remove WOODPECKER_DOCS config [#2647]
- Remove plugin-only option from secrets [#2213]
- Remove deprecated API paths [#2639]
- Remove SSH backend [#2635]
- Remove deprecated build command [#2602]
- Deprecate "platform" filter in favour of "labels" [#2181]
- Remove unused "sync" option from RepoListOpts from the client lib [#2090]
- Drop deprecated built-in environment variables [#2048]

### How we plan to handle releases in the future

### Migration notes

## New features

### Improved error and linter in the UI

### Reposiotry & organization lists in the admin UI

<https://github.com/woodpecker-ci/woodpecker/pull/2338>
<https://github.com/woodpecker-ci/woodpecker/pull/2347>

### Support for user secrets

### Bitbucket cloud support for multi-workflows

## Changelog

The full changelog can be viewed in our project source folder at [CHANGELOG.md](https://github.com/woodpecker-ci/woodpecker/blob/v1.0.0/CHANGELOG.md)
