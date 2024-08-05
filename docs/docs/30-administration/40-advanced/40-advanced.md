# Advanced options

Why should we be happy with a default setup? We should not! Woodpecker offers a lot of advanced options to configure it to your needs.

## Behind a proxy

See the [proxy guide](./10-proxy.md) if you want to see a setup behind Apache, Nginx, Caddy or ngrok.

In the case you need to use Woodpecker with a URL path prefix (like: <https://example.org/woodpecker/>), add the root path to [`WOODPECKER_HOST`](../10-server-config.md#woodpecker_host).

## SSL

Woodpecker supports SSL configuration by using Let's encrypt or by using own certificates. See the [SSL guide](./20-ssl.md).

## Metrics

A [Prometheus endpoint](./90-prometheus.md) is exposed by Woodpecker to collect metrics.

## Autoscaling

The [autoscaler](./30-autoscaler.md) can be used to deploy new agents to a cloud provider based on the current workload your server is experiencing.

## Configuration service

Sometime the normal yaml configuration compiler isn't enough. You can use the [configuration service](./100-external-configuration-api.md) to process your configuration files by your own.
