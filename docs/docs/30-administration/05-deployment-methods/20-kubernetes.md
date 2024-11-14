# Kubernetes

We recommended to deploy Woodpecker using the [Woodpecker helm chart](https://github.com/woodpecker-ci/helm).
Have a look at the [`values.yaml`](https://github.com/woodpecker-ci/helm/blob/main/charts/woodpecker/values.yaml) config files for all available settings.

The chart contains two sub-charts, `server` and `agent` which are automatically configured as needed.
The chart started off with two independent charts but was merged into one to simplify the deployment at start of 2023.

A couple of backend-specific config env vars exists which are described in the [kubernetes backend docs](../22-backends/40-kubernetes.md).

## Metrics

Please see [Prometheus](../40-advanced/90-prometheus.md) for general information on configuration and usage.

For Kubernetes, you must set the following values when deploying via Helm chart to enable in-cluster metrics gathering:

```yaml
metrics:
  enabled: true
  port: 9001
```

This activates the `/metrics` endpoint on port `9001` without authentication. This port is not exposed externally by default. Use the instructions at [Prometheus](../40-advanced/90-prometheus.md) if you want to enable authenticated external access to metrics.

To enable Prometheus pod monitoring discovery, you must also make the following settings:

<!-- cspell:disable -->

```yaml
prometheus:
  podmonitor:
    enabled: true
    interval: 60s
    labels: {}
```

<!-- cspell:enable -->

### Troubleshooting Metrics

If you are not receiving metrics despite the steps above, ensure that in your Prometheus configuration either your namespace is explicitly configured in `podMonitorNamespaceSelector` or the selectors are disabled.

```yaml
# Search all available namespaces
podMonitorNamespaceSelector:
  matchLabels: {}
# Enable all available pod monitors
podMonitorSelector:
  matchLabels: {}
```
