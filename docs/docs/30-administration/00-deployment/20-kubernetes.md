# Kubernetes

We recommended to deploy Woodpecker using the [Woodpecker helm chart](https://github.com/woodpecker-ci/helm).
Have a look at the [`values.yaml`](https://github.com/woodpecker-ci/helm/blob/main/values.yaml) config files for all available settings.

The chart contains two subcharts, `server` and `agent` which are automatically configured as needed.
The chart started off with two independent charts but was merged into one to simplify the deployment at start of 2023.

A couple of backend-specific config env vars exists which are described in the [kubernetes backend docs](../22-backends/40-kubernetes.md).
