# SSL

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
