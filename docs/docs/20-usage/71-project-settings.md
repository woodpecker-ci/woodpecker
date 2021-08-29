# Project settings

As the owner of a project in Woodpecker you can change some project related settings via the Webinterface.

![project settings](./project-settings.png)

## Pipeline path

The path to the pipeline file or folder. By default it point `.woodpecker.yml`. To use a [multi pipeline](/docs/usage/multi-pipeline) you have to change it to a folder path ending with `/` like `.woodpecker/`.

If you enable the fallback check, Woodpecker will first try to load the configuration from the defined path and if it fails to find that file it will try to use `.drone.yml`.

## Repository hooks

Your Version-Control-System will notify Woodpecker about some events via webhooks. If you want your pipeline to only run on specific webhooks, you can check them by this setting.

## Project settings

### Protected

Every build initiated by a user (not including the project owner) needs to be approved by the owner before being executed. This can be used if your repository is public to protect the pipeline configuration from running unauthorized changes on third-party pull requests.

### Trusted

If you set your project to trusted, a pipeline step and by this the underlying containers gets access to escalated capabilities like mounting volumes.

## Project visibility

You can change the visibility of your project by this setting. If a user has access to a project he can see all builds and their logs and artifacts. Settings, Secrets and Registries can only be accessed by owners.

- `Public` Every user can see your project without being logged in.
- `Private` Only authenticated users of the Woodpecker instance can see this project.
- `Internal` Only you and other owners of the repository can see this project.

## Timeout

After this timeout a pipeline has to finish or will be treated as timed out.

