# Extensions

Woodpecker allows you to replace internal logic with external extensions by using pre-defined http endpoints.

There are currently two types of extensions available:

- [Configuration extension](./40-configuration-extension.md) to modify or generate Woodpeckers pipeline configurations.

## Security

:::warning
You need to trust the extensions as they are receiving private information like secrets and tokens and might return harmful
data like malicious pipeline configurations that could be executed.
:::

To prevent your extensions from such attacks, Woodpecker is signing all http-requests using [http signatures](https://tools.ietf.org/html/draft-cavage-http-signatures). Woodpecker therefore uses a public-private ed25519 key pair. To verify the requests your extension has to verify the signature of all request using the public key with some library like [httpsig](https://github.com/yaronf/httpsign). You can get the public Woodpecker key by opening `http://my-woodpecker.tld/api/signature/public-key` or by visiting the Woodpecker UI, going to you repo settings and opening the extensions page.

## Example extensions

A simplistic service providing endpoints for a config and secrets extension can be found here: [https://github.com/woodpecker-ci/example-extensions](https://github.com/woodpecker-ci/example-extensions)

## Configuration

To prevent extensions from calling local services by default only external hosts / ip-addresses are allowed. You can change this behavior by setting the `WOODPECKER_ALLOWED_EXTENSIONS_HOSTS` environment variable. You can use a comma separated list of:

- Built-in networks:
  - `loopback`: 127.0.0.0/8 for IPv4 and ::1/128 for IPv6, localhost is included.
  - `private`: RFC 1918 (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) and RFC 4193 (FC00::/7). Also called LAN/Intranet.
  - `external`: A valid non-private unicast IP, you can access all hosts on public internet.
  - `*`: All hosts are allowed.
- CIDR list: `1.2.3.0/8` for IPv4 and `2001:db8::/32` for IPv6
- (Wildcard) hosts: `example.com`, `*.example.com`, `192.168.100.*`
