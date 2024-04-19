# Overview

## Supported features

| Feature | [GitHub](github/) | [Gitea / Forgejo](gitea/) | [Gitlab](gitlab/) | [Bitbucket](bitbucket/) |
| --- | :---: | :---: | :---: | :---: |
| Event: Push | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Event: Tag | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Event: Pull-Request | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Event: Deploy | :white_check_mark: | :x: | :x: | :x: |
| [Multiple workflows](../../20-usage/25-workflows.md) | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x: |
| [when.path filter](../../20-usage/20-pipeline-syntax.md#path) | :white_check_mark: | :white_check_mark:¹ | :white_check_mark: | :x: |

¹ for pull requests at least Gitea version 1.17 is required
