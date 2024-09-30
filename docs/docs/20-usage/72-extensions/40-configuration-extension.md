# Configuration extension

The configuration extension can be used to modify or generate Woodpeckers pipeline configurations. You can configure a HTTP
endpoint in the repository settings in the extensions tab.

Using such an extension can be useful if you want to:

- Preprocess the original configuration file with something like go templating
- Convert custom attributes to Woodpecker attributes
- Add defaults to the configuration like default steps
- Convert configuration files from a totally different format like Gitlab CI config, Starlark, Jsonnet, ...
- Centralize configuration for multiple repositories in one place

## Security

:::warning
As Woodpecker will pass private information like tokens and will execute the returned configuration, it is extremely important to secure the external extension. Therefore Woodpecker signs every request. Read more about it in the [security section](./10-extensions.md#security).
:::

## Global configuration

In addition to the ability to configure the extension per repository, you can also configure a global endpoint in the Woodpecker server configuration. This can be useful if you want to use the extension for all repositories. Be careful if
you share your Woodpecker server with others as they will also use your configuration extension.

The global configuration will be called before the repository specific configuration extension if both are configured.

```ini title="Server"
WOODPECKER_CONFIG_SERVICE_ENDPOINT=https://example.com/ciconfig
```

## How it works

When a pipeline is triggered Woodpecker will fetch the pipeline configuration from the repository, then make a HTTP POST request to the configured extension with a JSON payload containing some data like the repository, pipeline information and the current config files retrieved from the repository. The extension can then send back modified or even new pipeline configurations following Woodpeckers official yaml format that should be used.

:::tip
The netrc data is pretty powerful as it contains credentials to access the repository. You can use this to clone the repository or even use the forge (Github or Gitlab, ...) api to get more information about the repository.
:::

### Request

The extension receives an HTTP POST request with the following JSON payload:

```ts
class Request {
  repo: Repo;
  pipeline: Pipeline;
  netrc: Netrc;
  configuration: {
    name: string; // filename of the configuration file
    data: string; // content of the configuration file
  }[];
}
```

Checkout the following models for more information:

- [repo model](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/repo.go)
- [pipeline model](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/pipeline.go)
- [netrc model](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/netrc.go)

:::tip
The `netrc` data is pretty powerful as it contains credentials to access the repository. You can use this to clone the repository or even use the forge (Github or Gitlab, ...) api to get more information about the repository.
:::

Example request:

```json
{
  "repo": {
    "id": 100,
    "uid": "",
    "user_id": 0,
    "namespace": "",
    "name": "woodpecker-testpipe",
    "slug": "",
    "scm": "git",
    "git_http_url": "",
    "git_ssh_url": "",
    "link": "",
    "default_branch": "",
    "private": true,
    "visibility": "private",
    "active": true,
    "config": "",
    "trusted": false,
    "protected": false,
    "ignore_forks": false,
    "ignore_pulls": false,
    "cancel_pulls": false,
    "timeout": 60,
    "counter": 0,
    "synced": 0,
    "created": 0,
    "updated": 0,
    "version": 0
  },
  "pipeline": {
    "author": "myUser",
    "author_avatar": "https://myforge.com/avatars/d6b3f7787a685fcdf2a44e2c685c7e03",
    "author_email": "my@email.com",
    "branch": "main",
    "changed_files": ["somefilename.txt"],
    "commit": "2fff90f8d288a4640e90f05049fe30e61a14fd50",
    "created_at": 0,
    "deploy_to": "",
    "enqueued_at": 0,
    "error": "",
    "event": "push",
    "finished_at": 0,
    "id": 0,
    "link_url": "https://myforge.com/myUser/woodpecker-testpipe/commit/2fff90f8d288a4640e90f05049fe30e61a14fd50",
    "message": "test old config\n",
    "number": 0,
    "parent": 0,
    "ref": "refs/heads/main",
    "refspec": "",
    "clone_url": "",
    "reviewed_at": 0,
    "reviewed_by": "",
    "sender": "myUser",
    "signed": false,
    "started_at": 0,
    "status": "",
    "timestamp": 1645962783,
    "title": "",
    "updated_at": 0,
    "verified": false
  },
  "configs": [
    {
      "name": ".woodpecker.yaml",
      "data": "steps:\n  - name: backend\n    image: alpine\n    commands:\n      - echo \"Hello there from Repo (.woodpecker.yaml)\"\n"
    }
  ]
}
```

### Response

The extension should respond with a JSON payload containing the new configuration files in Woodpeckers official yaml format.
If the extension wants to keep the existing configuration files, it can respond with **HTTP 204**.

```ts
class Response {
  configs: {
    name: string; // filename of the configuration file
    data: string; // content of the configuration file
  }[];
}
```

Example response:

```json
{
  "configs": [
    {
      "name": "central-override",
      "data": "steps:\n  - name: backend\n    image: alpine\n    commands:\n      - echo \"Hello there from ConfigAPI\"\n"
    }
  ]
}
```
