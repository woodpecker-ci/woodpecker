# Proxy

## Apache

This guide provides a brief overview for installing Woodpecker server behind the Apache2 web-server. This is an example configuration:

<!-- cspell:ignore apacheconf -->

```apacheconf
ProxyPreserveHost On

RequestHeader set X-Forwarded-Proto "https"

ProxyPass / http://127.0.0.1:8000/
ProxyPassReverse / http://127.0.0.1:8000/
```

You must have these Apache modules installed:

- `proxy`
- `proxy_http`

You must configure Apache to set `X-Forwarded-Proto` when using https.

```diff
 ProxyPreserveHost On

+RequestHeader set X-Forwarded-Proto "https"

 ProxyPass / http://127.0.0.1:8000/
 ProxyPassReverse / http://127.0.0.1:8000/
```

## Nginx

This guide provides a basic overview for installing Woodpecker server behind the Nginx web-server. For more advanced configuration options please consult the official Nginx [documentation](https://docs.nginx.com/nginx/admin-guide).

Example configuration:

```nginx
server {
    listen 80;
    server_name woodpecker.example.com;

    location / {
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Host $http_host;

        proxy_pass http://127.0.0.1:8000;
        proxy_redirect off;
        proxy_http_version 1.1;
        proxy_buffering off;

        chunked_transfer_encoding off;
    }
}
```

You must configure the proxy to set `X-Forwarded` proxy headers:

```diff
 server {
     listen 80;
     server_name woodpecker.example.com;

     location / {
+        proxy_set_header X-Forwarded-For $remote_addr;
+        proxy_set_header X-Forwarded-Proto $scheme;

         proxy_pass http://127.0.0.1:8000;
         proxy_redirect off;
         proxy_http_version 1.1;
         proxy_buffering off;

         chunked_transfer_encoding off;
     }
 }
```

## Caddy

This guide provides a brief overview for installing Woodpecker server behind the [Caddy web-server](https://caddyserver.com/). This is an example caddyfile proxy configuration:

```caddy
# expose WebUI and API
woodpecker.example.com {
  reverse_proxy woodpecker-server:8000
}

# expose gRPC
woodpecker-agent.example.com {
  reverse_proxy h2c://woodpecker-server:9000
}
```

:::note
Above configuration shows how to create reverse-proxies for web and agent communication. If your agent uses SSL do not forget to enable [`WOODPECKER_GRPC_SECURE`](../15-agent-config.md#woodpecker_grpc_secure).
:::

## Tunnelmole

[Tunnelmole](https://github.com/robbie-cahill/tunnelmole-client) is an open source tunneling tool.

Start by [installing tunnelmole](https://github.com/robbie-cahill/tunnelmole-client#installation).

After the installation, run the following command to start tunnelmole:

```bash
tmole 8000
```

It will start a tunnel and will give a response like this:

```bash
➜  ~ tmole 8000
http://bvdo5f-ip-49-183-170-144.tunnelmole.net is forwarding to localhost:8000
https://bvdo5f-ip-49-183-170-144.tunnelmole.net is forwarding to localhost:8000
```

Set `WOODPECKER_HOST` to the Tunnelmole URL (`xxx.tunnelmole.net`) and start the server.

## Ngrok

[Ngrok](https://ngrok.com/) is a popular closed source tunnelling tool. After installing ngrok, open a new console and run the following command:

```bash
ngrok http 8000
```

Set `WOODPECKER_HOST` to the ngrok URL (usually xxx.ngrok.io) and start the server.

## Traefik

To install the Woodpecker server behind a [Traefik](https://traefik.io/) load balancer, you must expose both the `http` and the `gRPC` ports. Here is a comprehensive example, considering you are running Traefik with docker swarm and want to do TLS termination and automatic redirection from http to https.

<!-- cspell:words redirectscheme certresolver  -->

```yaml
services:
  server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_ADMIN=your_admin_user
      # other settings ...

    networks:
      - dmz # externally defined network, so that traefik can connect to the server
    volumes:
      - woodpecker-server-data:/var/lib/woodpecker/

    deploy:
      labels:
        - traefik.enable=true

        # web server
        - traefik.http.services.woodpecker-service.loadbalancer.server.port=8000

        - traefik.http.routers.woodpecker-secure.rule=Host(`cd.your-domain.com`)
        - traefik.http.routers.woodpecker-secure.tls=true
        - traefik.http.routers.woodpecker-secure.tls.certresolver=letsencrypt
        - traefik.http.routers.woodpecker-secure.entrypoints=web-secure
        - traefik.http.routers.woodpecker-secure.service=woodpecker-service

        - traefik.http.routers.woodpecker.rule=Host(`cd.your-domain.com`)
        - traefik.http.routers.woodpecker.entrypoints=web
        - traefik.http.routers.woodpecker.service=woodpecker-service

        - traefik.http.middlewares.woodpecker-redirect.redirectscheme.scheme=https
        - traefik.http.middlewares.woodpecker-redirect.redirectscheme.permanent=true
        - traefik.http.routers.woodpecker.middlewares=woodpecker-redirect@docker

        #  gRPC service
        - traefik.http.services.woodpecker-grpc.loadbalancer.server.port=9000
        - traefik.http.services.woodpecker-grpc.loadbalancer.server.scheme=h2c

        - traefik.http.routers.woodpecker-grpc-secure.rule=Host(`woodpecker-grpc.your-domain.com`)
        - traefik.http.routers.woodpecker-grpc-secure.tls=true
        - traefik.http.routers.woodpecker-grpc-secure.tls.certresolver=letsencrypt
        - traefik.http.routers.woodpecker-grpc-secure.entrypoints=web-secure
        - traefik.http.routers.woodpecker-grpc-secure.service=woodpecker-grpc

        - traefik.http.routers.woodpecker-grpc.rule=Host(`woodpecker-grpc.your-domain.com`)
        - traefik.http.routers.woodpecker-grpc.entrypoints=web
        - traefik.http.routers.woodpecker-grpc.service=woodpecker-grpc

        - traefik.http.middlewares.woodpecker-grpc-redirect.redirectscheme.scheme=https
        - traefik.http.middlewares.woodpecker-grpc-redirect.redirectscheme.permanent=true
        - traefik.http.routers.woodpecker-grpc.middlewares=woodpecker-grpc-redirect@docker

volumes:
  woodpecker-server-data:
    driver: local

networks:
  dmz:
    external: true
```

You should pass `WOODPECKER_GRPC_SECURE=true` and `WOODPECKER_GRPC_VERIFY=true` to your agent when using this configuration.
