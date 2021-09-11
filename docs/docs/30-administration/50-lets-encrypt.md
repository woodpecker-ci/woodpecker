# Let's Encrypt

Woodpecker supports automated SSL configuration and updates using Let's Encrypt.

You can enable Let's Encrypt by making the following modifications to your server configuration:

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    ports:
+     - 80:80
+     - 443:443
      - 9000:9000
    volumes:
      - /var/lib/drone:/var/lib/drone/
    restart: always
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
+     - WOODPECKER_LETS_ENCRYPT=true
```

Note that Woodpecker uses the hostname from the `WOODPECKER_HOST` environment variable when requesting certificates. For example, if `WOODPECKER_HOST=https://foo.com` the certificate is requested for `foo.com`.

>Once enabled you can visit your website at both the http and the https address

## Certificate Cache

Woodpecker writes the certificates to the below directory:

```
/var/lib/drone/golang-autocert
```

## Certificate Updates

Woodpecker uses the official Go acme library which will handle certificate upgrades. There should be no addition configuration or management required.
