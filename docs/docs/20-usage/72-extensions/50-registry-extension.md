# Registry extension

Woodpecker uses the registry extension to get registry credentials. You can configure an HTTP endpoint in the repository settings in the extensions tab.

Using such an extension can be useful if you want to:

- Centralize registry credential management
- Use an external storage for credentials
- Dynamically manage which credentials Woodpecker should use

## Security

:::warning
As Woodpecker will pass private information like tokens and will execute the returned configuration, it is extremely important to secure the external extension. Therefore Woodpecker signs every request. Read more about it in the [security section](./index.md#security).
:::

## Global configuration

In addition to the ability to configure the extension per repository, you can also configure a global endpoint in the Woodpecker server configuration. This can be useful if you want to use the extension for all repositories. Be careful if
you share your Woodpecker server with others as they will also use your registry extension.

If both the global and the repo-level extension return credentials for a registry, it will use the credentials from the repo extension.

```ini title="Server"
WOODPECKER_REGISTRY_SERVICE_ENDPOINT=https://example.com/ciconfig
```

## How it works

When a pipeline is triggered, Woodpecker will fetch the credentials from your service. As fallback, it uses the credentials configured directly in Woodpecker.

### Request

The extension receives an HTTP POST request with the following JSON payload:

```ts
class Request {
  repo: Repo;
  pipeline: Pipeline;
}
```

Checkout the following models for more information:

- [repo model](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/repo.go)
- [pipeline model](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/pipeline.go)

Example request:

```json
// Please check the latest structure in the models mentioned above.
// This example is likely outdated.

{
  "repo": {
    "id": 100,
    "uid": "",
    "user_id": 0,
    "namespace": "",
    "name": "woodpecker-test-pipeline",
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
    "changed_files": ["some-filename.txt"],
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
  }
}
```

### Response

The extension should respond with a JSON payload containing the new configuration files in Woodpecker's official YAML format.
If the extension wants to keep the existing configuration files, it can respond with HTTP status `204 No Content`.

```ts
class Response {
  registries: {
    address: string; // the docker registry address
    username: string; // registry username
    password: string; // registry password
  }[];
}
```

Example response:

```json
{
  "registries": [
    {
      "address": "docker.io",
      "username": "woodpecker-bot",
      "password": "your-pass-word-123"
    }
  ]
}
```
