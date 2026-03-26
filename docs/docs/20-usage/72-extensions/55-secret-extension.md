# Secret extension

Woodpecker uses the secret extension to get secrets from an external service. You can configure an HTTP endpoint in the repository settings in the extensions tab.

Using such an extension can be useful if you want to:

- Centralize secret management (e.g. HashiCorp Vault, AWS Secrets Manager)
- Dynamically generate secrets per pipeline

## Security

:::warning
As Woodpecker will pass private information like tokens and will execute the returned configuration, it is extremely important to secure the external extension. Therefore Woodpecker signs every request. Read more about it in the security section.
:::

## Global configuration

In addition to the ability to configure the extension per repository, you can also configure a global endpoint in the Woodpecker server configuration. This can be useful if you want to use the extension for all repositories. Be careful if
you share your Woodpecker server with others as they will also use your secret extension.

If both the global and the repo-level extension return a secret with the same name, it will use the secret from the repo extension.

```ini title="Server"
WOODPECKER_SECRET_EXTENSION_ENDPOINT=https://example.com/secrets
WOODPECKER_SECRET_EXTENSION_NETRC=false
```

## How it works

When a pipeline is triggered, Woodpecker will fetch secrets from your service. The extension secrets are merged with the secrets configured directly in Woodpecker, with extension secrets taking priority by name. If the extension is unavailable, Woodpecker falls back to the locally configured secrets.

### Request

The extension receives an HTTP POST request with the following JSON payload:

:::info
The `netrc` field is only included in the request when the global `WOODPECKER_SECRET_EXTENSION_NETRC` is set to `true` (default: `false`) or the per-repo "Send netrc credentials" is checked.
:::

```ts
class Request {
  repo: Repo;
  pipeline: Pipeline;
  netrc?: Netrc; // only included when netrc sending is enabled (see above)
}
```

Checkout the following models for more information:

- [repo model](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/repo.go)
- [pipeline model](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/pipeline.go)
- [netrc model](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/netrc.go)

:::tip
The `netrc` data is pretty powerful as it contains credentials to access the repository. You can use this to clone the repository or even use the forge (Github or Gitlab, ...) API to get more information about the repository.
:::

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
  },
  "netrc": {
    "machine": "myforge.com",
    "login": "myUser",
    "password": "forge-access-token"
  }
}
// Note: the "netrc" field is omitted when netrc sending is not enabled.
```

### Response

The extension should respond with a JSON object containing a `secrets` array.
If the extension wants to keep the existing secrets without adding any, it can respond with HTTP status `204 No Content`.

```ts
class Response {
  secrets: {
    name: string; // the secret name, matched by from_secret in pipeline config
    value: string; // the secret value
    images?: string[]; // optional: restrict to specific plugins
    events?: string[]; // optional: restrict to specific pipeline events
  }[];
}
```

Example response:

```json
{
  "secrets": [
    {
      "name": "docker_password",
      "value": "your-secret-password-123"
    },
    {
      "name": "deploy_token",
      "value": "super-secret-token",
      "events": ["push", "tag"]
    }
  ]
}
```
