# Kubernetes

We recommended to deploy Woodpecker using the [Woodpecker helm chart](https://github.com/woodpecker-ci/helm).
Have a look at the [`values.yaml`](https://github.com/woodpecker-ci/helm/blob/main/charts/woodpecker/values.yaml) config files for all available settings.

The chart contains two sub-charts, `server` and `agent` which are automatically configured as needed.
The chart started off with two independent charts but was merged into one to simplify the deployment at start of 2023.

A couple of backend-specific config env vars exists which are described in the [kubernetes backend docs](../22-backends/40-kubernetes.md).

## Metrics

Please see [Prometheus](../40-advanced/90-prometheus.md) for general configuration and usage information.

For Kubernetes, when deployed via Helm chart you will want to set the following values to enable in-cluster metrics gathering:

```yaml
  metrics:
    enabled: true
    port: 9001
```

This will enable /metrics on port :9001 without authentication.  This port is not externally exposed by default, use the instructions at [Prometheus](../40-advanced/90-prometheus.md) if you want to enable authenticated external access to metrics.

To enable Prometheus pod monitoring discovery, also set the following:

```yaml
  prometheus:
    podmonitor:
      enabled: true
      interval: 60s
      labels: {}
```

### Troubleshooting Metrics

If you are not receiving metrics despite doing the above, ensure your Prometheus configuration either has your namespace configured explicitly in `podMonitorNamespaceSelector`, or something similar to the following:

```yaml
    # Search all available namespaces
    podMonitorNamespaceSelector:
      matchLabels: {}
    # Enable all available pod monitors
    podMonitorSelector:
      matchLabels: {}
```
