# Extensions

Woodpecker allows you to replace internal logic with external extensions by using pre-defined http endpoints.

There are currently two types of extensions available:

- [Configuration extension](./40-configuration-extension.md) to modify or generate Woodpeckers pipeline configurations.
- [Secrets extension (alpha state)](./20-secrets-extension.md) to receive and update secrets from an external system like Hashicorp Vault or AWS Secrets Manager.

## Security

:::warning
You need to trust the extensions as they are receiving private information like secrets and tokens and might return harmful
data like malicious pipeline configurations that could be executed.
:::

To prevent your extensions from such attaks, Woodpecker is signing all http-requests using [http signatures](https://tools.ietf.org/html/draft-cavage-http-signatures). Woodpecker therefore uses a public-private ed25519 key pair. To verify the requests your extension has to verify the signature of all request using the public key with some library like [httpsig](https://github.com/go-fed/httpsig). You can get the public Woodpecker key by opening `http://my-woodpecker.tld/api/signature/public-key` or by visiting the Woodpecker UI, going to you repo settings and opening the extensions page.

## Example extensions

A simplistic serivce providing endpoints for a config and secrets extension can be found here: [https://github.com/woodpecker-ci/example-extensions](https://github.com/woodpecker-ci/example-extensions)
