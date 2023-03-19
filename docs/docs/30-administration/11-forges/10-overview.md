# Overview

## Supported features

| Feature | [GitHub](github/) | [Gitea](gitea/) | [Gitlab](gitlab/) | [Bitbucket](bitbucket/) | [Bitbucket Server](bitbucket_server/) | [Gogs](gogs/) |
| --- | :---: | :---: | :---: | :---: | :---: | :---: |
| Event: Push | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Event: Tag | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x: | :white_check_mark: | :white_check_mark: |
| Event: Pull-Request | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x: | :white_check_mark: |
| Event: Deploy | :white_check_mark: | :x: | :x: | :x: | :x: | :x: |
| OAuth | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x: |
| [Multiple workflows](../../20-usage/25-workflows.md) | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x: | :x: | :x: |
| [when.path filter](../../20-usage/20-pipeline-syntax.md#path) | :white_check_mark: | :white_check_mark:¹ | :white_check_mark: | :x: | :x: | :x: |

¹ for Gitea versions 1.17 or lower not for pull requests
