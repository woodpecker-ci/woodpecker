---
title: It's time for some changes - Woodpecker 2.1.0
description:
slug: release-v200
authors:
  - name: Anbraten
    title: Maintainer of Woodpecker
    url: https://github.com/anbraten
    image_url: https://github.com/anbraten.png
tags: [release, minor]
hide_table_of_contents: false
---

TODO

We want to say thanks to all our backers!!!
opencollective.com/woodpecker-

and also have a after x-mas/new year #present :
v2.1.0 is released: github.com/woodpecker-ci/woodp

#WoodpeckerCI #release

<!--truncate-->

## Features

### Pull request closed event

You ever wanted to shutdown a review environment when a pull request is closed? Now you can! We added a new event type `pull_request_closed` which is triggered when a pull request is closed or merged.

### Direct acyclic graph (DAG) support

Your step should run even after some previous steps finished, but you want to break the sequential execution? Now you can! We added support for DAGs, allowing you to define dependencies between steps.

## Changelog

The full changelog can be viewed in our project source folder at [CHANGELOG.md](https://github.com/woodpecker-ci/woodpecker/blob/v2.1.0/CHANGELOG.md)
