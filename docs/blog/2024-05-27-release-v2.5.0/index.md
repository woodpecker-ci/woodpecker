---
title: Here is Woodpecker 2.5.0
description: Introducing Woodpecker 2.5.0
slug: release-v250
authors:
  - name: Anbraten
    title: Maintainer of Woodpecker
    url: https://github.com/anbraten
    image_url: https://github.com/anbraten.png
tags: [release, minor]
hide_table_of_contents: false
---

Here is the next minor release 2.5.0 of Woodpecker 🪶 ☀️.

<!--truncate-->

As always thanks to all contributors who helped to make this release possible. It includes quite a few enhancements
most users will benefit from while they are probably not that visible at first sight for most. The release also includes some preparations for new features to come in the next versions. Anyway, let's dive into some of the highlights of this release:

## Improve the way entrypoints work

The implementation wasn't perfect yet so we improved the way entrypoints work:

If you define [`commands`](/docs/usage/workflow-syntax#commands), the default entrypoint will be `["/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"]`.

If you define your own entrypoint, you can completely overwrite the default entrypoint. If you define `entrypoint: ["/bin/my-script", ""]` for example you can run your own binary / script. In this case the commands section will ignored, however you can still access it in your own script by using the base64 encoded string of the `CI_SCRIPT` environment variable.

[#3269](https://github.com/woodpecker-ci/woodpecker/pull/3269)

## Cli output formats

The cli output has been improved. The first command (mainly pipeline info, ls, create) support a `--output` flag now which allows you to change the output format. There is a new `table` format (the new default) which will look like the following and can be further customized:

```bash
# use default table output
❯ woodpecker-cli pipeline ls --limit 2 2
NUMBER  STATUS   EVENT   BRANCH  COMMIT                                    AUTHOR
43      error    manual  main    473761d8b26b20f7c206408563d54cf998410329  woodpecker
42      success  push    main    473761d8b26b20f7c206408563d54cf998410329  woodpecker

# customize table output and disable header
❯ woodpecker-cli pipeline ls --limit 2 --output table=number,status,event --no-header 2
43  error    manual
42  success  push
```

In addition especially useful for programmatic usage there is a `go-template` output format which will output the data using the provided go template like this:

```bash
########
# go crazy and use a template layout
❯ woodpecker-cli pipeline ls --limit 2 --output go-template='{{range .}}{{printf "\x1b[33mPipeline #%d\x1b[0m\nStatus: %s\nEvent:%s\nCommit:%s\n\n" .Number .Status .Event .Commit}}{{end}}' 2
Pipeline #43
Status: error
Event:manual
Commit:473761d8b26b20f7c206408563d54cf998410329

Pipeline #42
Status: success
Event:push
Commit:473761d8b26b20f7c206408563d54cf998410329
```

[#3660](https://github.com/woodpecker-ci/woodpecker/pull/3660)

## Deleting logs or complete pipelines

If you accidentally exposed some secret to the public in your logs or you simply want to cleanup some logs you can now delete logs or complete pipelines using the api and the cli.

[#3451](https://github.com/woodpecker-ci/woodpecker/pull/3451)
[#3506](https://github.com/woodpecker-ci/woodpecker/pull/3506)
[#3458](https://github.com/woodpecker-ci/woodpecker/pull/3458)

## Support for Github deploy tasks

Woodpecker now supports Github deploy tasks. This allows you to pass the deploy task set in Github to your Woodpecker pipeline.

[#3512](https://github.com/woodpecker-ci/woodpecker/pull/3512)

## Deprecations

To keep things clean and simple we deprecated some pipeline options, server settings and features which will
be removed in the next major release:

- Deprecated `environment` filter, use `when.evaluate`
- Use `WOODPECKER_EXPERT_FORGE_OAUTH_HOST` instead of `WOODPECKER_DEV_GITEA_OAUTH_URL` or `WOODPECKER_DEV_OAUTH_HOST`
- Deprecated `WOODPECKER_WEBHOOK_HOST` in favor of `WOODPECKER_EXPERT_WEBHOOK_HOST`

For a full list of deprecations that will be dropped in the `next` major release `3.0.0` (no eta yet), please check the [migrations](/migrations#next) section.
