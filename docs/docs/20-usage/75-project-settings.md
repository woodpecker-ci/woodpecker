# Project settings

As the owner of a project in Woodpecker you can change project related settings via the web interface.

![project settings](./project-settings.png)

## Pipeline path

The path to the pipeline config file or folder. By default it is left empty which will use the following configuration resolution `.woodpecker/*.{yaml,yml}` -> `.woodpecker.yaml` -> `.woodpecker.yml`. If you set a custom path Woodpecker tries to load your configuration or fails if no configuration could be found at the specified location. To use a [multiple workflows](./25-workflows.md) with a custom path you have to change it to a folder path ending with a `/` like `.woodpecker/`.

## Repository hooks

Your Version-Control-System will notify Woodpecker about events via webhooks. If you want your pipeline to only run on specific webhooks, you can check them with this setting.

## Allow pull requests

Enables handling webhook's pull request event. If disabled, then pipeline won't run for pull requests.

## Allow deployments

Enables a pipeline to be started with the `deploy` event from a successful pipeline.

:::danger
Only activate this option if you trust all users who have push access to your repository.
Otherwise, these users will be able to steal secrets that are only available for `deploy` events.
:::

## Require approval for

To prevent malicious pipelines from extracting secrets or running harmful commands or to prevent accidental pipeline runs, you can require approval for an additional review process. Depending on the enabled option, a pipeline will be put on hold after creation and will only continue after approval. The default restrictive setting is `Approvals for forked repositories`.

## Trusted

If you set your project to trusted, a pipeline step and by this the underlying containers gets access to escalated capabilities like mounting volumes.

:::note

Only server admins can set this option. If you are not a server admin this option won't be shown in your project settings.

:::

## Custom trusted clone plugins

The clone step may require Git credentials (e.g. for private repos) which are injected via `netrc`.

They are only injected into trusted clone plugins listed in the env var `WOODPECKER_PLUGINS_TRUSTED_CLONE` or in this repo setting.

## Project visibility

You can change the visibility of your project by this setting. If a user has access to a project they can see all builds and their logs and artifacts. Settings, Secrets and Registries can only be accessed by owners.

- `Public` Every user can see your project without being logged in.
- `Internal` Only authenticated users of the Woodpecker instance can see this project.
- `Private` Only you and other owners of the repository can see this project.

## Timeout

After this timeout a pipeline has to finish or will be treated as timed out.

## Cancel previous pipelines

By enabling this option for a pipeline event previous pipelines of the same event and context will be canceled before starting the newly triggered one.
