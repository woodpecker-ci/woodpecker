# Proxy

## Apache

This guide provides a brief overview for installing Woodpecker server behind the Apache2 webserver. This is an example configuration:

```nohighlight
ProxyPreserveHost On

RequestHeader set X-Forwarded-Proto "https"

ProxyPass / http://127.0.0.1:8000/
ProxyPassReverse / http://127.0.0.1:8000/
```

You must have the below Apache modules installed.

```nohighlight
a2enmod proxy
a2enmod proxy_http
```

You must configure Apache to set `X-Forwarded-Proto` when using https.

```diff
ProxyPreserveHost On

+RequestHeader set X-Forwarded-Proto "https"

ProxyPass / http://127.0.0.1:8000/
ProxyPassReverse / http://127.0.0.1:8000/
```

## Nginx

This guide provides a basic overview for installing Woodpecker server behind the nginx webserver. For more advanced configuration options please consult the official nginx [documentation](https://www.nginx.com/resources/admin-guide/).

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
+       proxy_set_header X-Forwarded-For $remote_addr;
+       proxy_set_header X-Forwarded-Proto $scheme;

        proxy_pass http://127.0.0.1:8000;
        proxy_redirect off;
        proxy_http_version 1.1;
        proxy_buffering off;

        chunked_transfer_encoding off;
    }
}
```

## Caddy

This guide provides a brief overview for installing Woodpecker server behind the [Caddy webserver](https://caddyserver.com/). This is an example caddyfile proxy configuration:

```nohighlight
woodpecker.example.com {
  reverse_proxy woodpecker-server:8000
}
```

## Ngrok

After installing [ngrok](https://ngrok.com/), open a new console and run:

```sh
ngrok http 8000
```

Set `WOODPECKER_HOST` (for example in `docker-compose.yml`) to the ngrok url (usually xxx.ngrok.io) and start the server.
