# Helm Chart

Woodpecker provides a [Helm chart](https://github.com/woodpecker-ci/helm) for Kubernetes environments:

```bash
helm install woodpecker oci://ghcr.io/woodpecker-ci/helm/woodpecker
```

## Configuration

To fetch all configurable options with detailed comments:

```bash
helm show values oci://ghcr.io/woodpecker-ci/helm/woodpecker > values.yaml
```

Install using custom values:

```bash
helm install woodpecker \
  oci://ghcr.io/woodpecker-ci/helm/woodpecker \
  -f values.yaml
```

### Metrics

To enable metrics gathering, set the following in values.yml:

```yaml
metrics:
  enabled: true
  port: 9001
```

This activates the `/metrics` endpoint on port `9001` without authentication. This port is not exposed externally by default. Use the instructions at Prometheus if you want to enable authenticated external access to metrics.

To enable both Prometheus pod monitoring discovery, set:

<!-- cspell:disable -->

```yaml
prometheus:
  podmonitor:
    enabled: true
    interval: 60s
    labels: {}
```

<!-- cspell:enable -->

If you are not receiving metrics after following the steps above, verify that your Prometheus configuration includes your namespace explicitly in the podMonitorNamespaceSelector or that the selectors are disabled:

```yaml
# Search all available namespaces
podMonitorNamespaceSelector:
  matchLabels: {}
# Enable all available pod monitors
podMonitorSelector:
  matchLabels: {}
```
