# Getting started

## Repository Activation

To activate your project navigate to your account settings. You will see a list of repositories which can be activated with a simple toggle. When you activate your repository, Woodpecker automatically adds webhooks to your forge (e.g. GitHub).

Webhooks are used to trigger pipeline executions. When you push code to your repository, open a pull request, or create a tag, your forge will automatically send a webhook to Woodpecker which will in turn trigger pipeline execution.

![repository list](repo-list.png)

> Required Permissions
>
>The user who enables a repo in Woodpecker must have `Admin` rights on that repo, so that Woodpecker can add the webhook.
>
> Note that manually creating webhooks yourself is not possible. This is because webhooks are signed using a per-repository secret key which is not exposed to end users.

# Webhooks

When you activate your repository Woodpecker automatically add webhooks to your forge (e.g. GitHub). There is no manual configuration required.

Webhooks are used to trigger pipeline executions. When you push code to your repository, open a pull request, or create a tag, your forge will automatically send a webhook to Woodpecker which will in turn trigger pipeline execution.

## Configuration

To configure your pipeline you should place a `.woodpecker.yml` file in the root of your repository. The .woodpecker.yml file is used to define your pipeline steps. It is a superset of the widely used docker-compose file format.

:::note
We support most of YAML 1.2, but preserve some behavior from 1.1 for backward compatibility.
Read more at: [https://github.com/go-yaml/yaml](https://github.com/go-yaml/yaml/tree/v3)
:::

Example pipeline configuration:

```yaml
steps:
  build:
    image: golang
    commands:
      - go get
      - go build
      - go test

services:
  postgres:
    image: postgres:9.4.5
    environment:
      - POSTGRES_USER=myapp
```

Example pipeline configuration with multiple, serial steps:

```yaml
steps:
  backend:
    image: golang
    commands:
      - go get
      - go build
      - go test

  frontend:
    image: node:6
    commands:
      - npm install
      - npm test

  notify:
    image: plugins/slack
    channel: developers
    username: woodpecker
```

## Execution

To trigger your first pipeline execution you can push code to your repository, open a pull request, or push a tag. Any of these events triggers a webhook from your forge and execute your pipeline.
