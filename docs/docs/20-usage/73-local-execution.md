# Local pipeline execution

`woodpecker-cli exec` runs workflow files from your local checkout. Use it to test pipeline changes before pushing them, to debug a workflow without waiting for a server run, or to replay a server pipeline with downloaded metadata.

## Requirements

- Install `woodpecker-cli` from the [distribution packages](../30-administration/05-installation/30-packages.md) or a release archive.
- Run the command from the repository checkout, or pass `--repo-path` to point at it.
- Make sure the backend you want to use is available locally. The Docker backend needs access to a Docker daemon. The local backend runs commands directly on your host and does not reproduce the container image environment.

## Run a workflow file

Create or edit a workflow file, then run it directly:

```shell
woodpecker-cli exec .woodpecker/my-first-workflow.yaml
```

You can also run every `.yaml` and `.yml` file in a workflow directory:

```shell
woodpecker-cli exec .woodpecker/
```

By default, Woodpecker auto-detects a backend. Select one explicitly when you want the local run to match a specific agent backend:

```shell
woodpecker-cli exec --backend-engine docker .woodpecker/my-first-workflow.yaml
woodpecker-cli exec --backend-engine local .woodpecker/my-first-workflow.yaml
```

## Pass metadata

Metadata values are set automatically, but you can override them to test conditions such as branches, pull requests, tags, and events:

```shell
woodpecker-cli exec \
  --pipeline-event push \
  --commit-branch main \
  --commit-sha "$(git rev-parse HEAD)" \
  --repo octocat/hello-world \
  .woodpecker/my-first-workflow.yaml
```

If you downloaded pipeline metadata from the Woodpecker UI, pass it with `--metadata-file` and adjust individual values with other flags when needed:

```shell
woodpecker-cli exec \
  --metadata-file pipeline-metadata.json \
  --pipeline-event pull_request \
  .woodpecker/my-first-workflow.yaml
```

:::warning
The metadata file is not a stable, portable API: its format is only guaranteed for the same server and CLI version it came from. Use it to replay against a matching version, and re-download after upgrades instead of reusing old files.
:::

## Use a configuration extension

If your server gets the workflows from a [configuration extension](./72-extensions/40-configuration-extension.md), pass the same endpoint to run the workflows the server would run:

```shell
woodpecker-cli exec \
  --config-extension-endpoint https://example.com/ciconfig \
  --repo octocat/hello-world \
  --commit-branch main
```

The CLI sends the request the server sends, built from the metadata flags. Workflow files given as arguments, or found at the default config paths, are sent as `configuration` and are run when the extension responds with `204 No Content`. Without local files the request has no `configuration`, as in exclusive mode.

The request is signed the same way the server signs extension calls: an ed25519 [HTTP signature](https://tools.ietf.org/html/draft-cavage-http-signatures) over `@request-target` and `content-digest`. The signature name is `woodpecker-cli-extensions`, not the server's `woodpecker-ci-extensions`, so the extension can trust both. The first time the option is used, the private key is generated and stored in the system keyring, in the same place as CLI tokens (`woodpecker-cli` / `config-extension-signing-key`). The public key is printed as PEM, in the same format as the server's `GET /api/signature/public-key`. Configure the extension with that public key.

:::note
The CLI does not send `netrc` credentials. A system without a keyring cannot store the signing key, and the command fails instead of sending an unsigned request.
:::

## Pass environment variables and secrets

Use `--env` for regular environment variables:

```shell
woodpecker-cli exec \
  --env GOFLAGS=-mod=readonly \
  .woodpecker/test.yaml
```

Secrets are not downloaded from the server. Pass the values needed for local debugging explicitly:

```shell
woodpecker-cli exec \
  --secrets deploy_token="$DEPLOY_TOKEN" \
  .woodpecker/deploy.yaml
```

For multiple secrets, keep them in a local YAML file that is ignored by Git:

```yaml title=".woodpecker/local-secrets.yaml"
deploy_token: ghp_example
registry_password: example-password
```

```shell
woodpecker-cli exec \
  --secrets-file .woodpecker/local-secrets.yaml \
  .woodpecker/deploy.yaml
```

## More options

See the generated [CLI reference](../40-cli/90-reference.md#exec) for the full list of `exec` flags.
