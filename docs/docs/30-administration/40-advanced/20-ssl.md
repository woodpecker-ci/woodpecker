# SSL

Woodpecker supports two ways of enabling SSL communication. You can either use Let's Encrypt to get automated SSL support with
renewal or provide your own SSL certificates.

## Let's Encrypt

Woodpecker supports automated SSL configuration and updates using Let's Encrypt.

You can enable Let's Encrypt by making the following modifications to your server configuration:

```ini
WOODPECKER_LETS_ENCRYPT=true
WOODPECKER_LETS_ENCRYPT_EMAIL=ssl-admin@example.tld
```

Note that Woodpecker uses the hostname from the `WOODPECKER_HOST` environment variable when requesting certificates. For example, if `WOODPECKER_HOST=https://example.com` is set the certificate is requested for `example.com`. To receive emails before certificates expire Let's Encrypt requires an email address. You can set it with `WOODPECKER_LETS_ENCRYPT_EMAIL=ssl-admin@example.tld`.

The SSL certificates are stored in `$HOME/.local/share/certmagic` for binary versions of Woodpecker and in `/var/lib/woodpecker` for the Container versions of it. You can set a custom path by setting `XDG_DATA_HOME` if required.

> Once enabled you can visit the Woodpecker UI with http and the HTTPS address. HTTP will be redirected to HTTPS.

### Certificate Cache

Woodpecker writes the certificates to `/var/lib/woodpecker/certmagic/`.

### Certificate Updates

Woodpecker uses the official Go acme library which will handle certificate upgrades. There should be no addition configuration or management required.

## SSL with own certificates

Woodpecker supports SSL configuration by mounting certificates into your container.

```ini
WOODPECKER_SERVER_CERT=/etc/certs/woodpecker.example.com/server.crt
WOODPECKER_SERVER_KEY=/etc/certs/woodpecker.example.com/server.key
```

### Certificate Chain

The most common problem encountered is providing a certificate file without the intermediate chain.

> LoadX509KeyPair reads and parses a public/private key pair from a pair of files. The files must contain PEM encoded data. The certificate file may contain intermediate certificates following the leaf certificate to form a certificate chain.

### Certificate Errors

SSL support is provided using the [ListenAndServeTLS](https://golang.org/pkg/net/http/#ListenAndServeTLS) function from the Go standard library. If you receive certificate errors or warnings please examine your configuration more closely.

### Running in containers

Update your configuration to expose the following ports:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     ports:
+      - 80:80
+      - 443:443
       - 9000:9000
```

Update your configuration to mount your certificate and key:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     volumes:
+      - /etc/certs/woodpecker.example.com/server.crt:/etc/certs/woodpecker.example.com/server.crt
+      - /etc/certs/woodpecker.example.com/server.key:/etc/certs/woodpecker.example.com/server.key
```

Update your configuration to provide the paths of your certificate and key:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
+      - WOODPECKER_SERVER_CERT=/etc/certs/woodpecker.example.com/server.crt
+      - WOODPECKER_SERVER_KEY=/etc/certs/woodpecker.example.com/server.key
```
