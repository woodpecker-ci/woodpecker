# Overview

## Supported features

| Feature | [GitHub](github/) | [Gitea](gitea/) | [Gitlab](gitlab/) | [Bitbucket](bitbucket/) | [Bitbucket Server](bitbucket_server/) | [Gogs](gogs/) | [Coding](coding/) |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| Event: Push | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Event: Tag | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Event: Pull-Request | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Event: Deploy | :white_check_mark: | :x: | :x: |
| OAuth | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| [Multi pipeline](../../20-usage/25-multi-pipeline.md) | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x: | :x: | :x: | :x: |
| [when.path filter](../../20-usage/22-conditional-execution.md#path) | :white_check_mark: | :white_check_mark:¹ | :white_check_mark: | :x: | :x: | :x: | :x: |

¹) [except for pull requests](https://github.com/woodpecker-ci/woodpecker/issues/754)
