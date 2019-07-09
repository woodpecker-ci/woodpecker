Drone supports ssl configuration by mounting certificates into your container.

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    ports:
+     - 80:80
+     - 443:443
      - 9000:9000
    volumes:
      - /var/lib/drone:/var/lib/drone/
+     - /etc/certs/drone.foo.com/server.crt:/etc/certs/drone.foo.com/server.crt
+     - /etc/certs/drone.foo.com/server.key:/etc/certs/drone.foo.com/server.key
    restart: always
    environment:
+     - DRONE_SERVER_CERT=/etc/certs/drone.foo.com/server.crt
+     - DRONE_SERVER_KEY=/etc/certs/drone.foo.com/server.key
```

Update your configuration to expose the following ports:

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    ports:
+     - 80:80
+     - 443:443
      - 9000:9000
```

Update your configuration to mount your certificate and key:

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    ports:
      - 80:80
      - 443:443
      - 9000:9000
    volumes:
      - /var/lib/drone:/var/lib/drone/
+     - /etc/certs/drone.foo.com/server.crt:/etc/certs/drone.foo.com/server.crt
+     - /etc/certs/drone.foo.com/server.key:/etc/certs/drone.foo.com/server.key
```

Update your configuration to provide the paths of your certificate and key:

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    ports:
      - 80:80
      - 443:443
      - 9000:9000
    volumes:
      - /var/lib/drone:/var/lib/drone/
      - /etc/certs/drone.foo.com/server.crt:/etc/certs/drone.foo.com/server.crt
      - /etc/certs/drone.foo.com/server.key:/etc/certs/drone.foo.com/server.key
    restart: always
    environment:
+     - DRONE_SERVER_CERT=/etc/certs/drone.foo.com/server.crt
+     - DRONE_SERVER_KEY=/etc/certs/drone.foo.com/server.key
```

# Certificate Chain

The most common problem encountered is providing a certificate file without the intermediate chain.

> LoadX509KeyPair reads and parses a public/private key pair from a pair of files. The files must contain PEM encoded data. The certificate file may contain intermediate certificates following the leaf certificate to form a certificate chain.

# Certificate Errors

SSL support is provided using the [ListenAndServeTLS](https://golang.org/pkg/net/http/#ListenAndServeTLS) function from the Go standard library. If you receive certificate errors or warnings please examine your configuration more closely. Please do not create issues claiming SSL is broken.
