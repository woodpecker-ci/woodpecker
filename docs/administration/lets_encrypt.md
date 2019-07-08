Drone supports automated ssl configuration and updates using let's encrypt. You can enable let's encrypt by making the following modifications to your server configuration:

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
    restart: always
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
+     - DRONE_LETS_ENCRYPT=true
```

Note that Drone uses the hostname from the `DRONE_HOST` environment variable when requesting certificates. For example, if `DRONE_HOST=https://foo.com` the certificate is requested for `foo.com`.

>Once enabled you can visit your website at both the http and the https address

# Certificate Cache

Drone writes the certificates to the below directory:

```
/var/lib/drone/golang-autocert
```

# Certificate Updates

Drone uses the official Go acme library which will handle certificate upgrades. There should be no addition configuration or management required.
