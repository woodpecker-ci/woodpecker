# Kubernetes

Woodpecker does support Kubernetes as a backend.

:::caution
Kubernetes support is still experimental and not all pipeline features are fully supported yet.

Check the [current state](https://github.com/woodpecker-ci/woodpecker/issues/9#issuecomment-483979755)
:::

## Deploy with HELM

Deploying Woodpecker with [HELM](https://helm.sh/docs/) is the recommended way.
Have a look at the `values.yaml` config files for all available settings.

### Preparation

```shell
# create agent secret
kubectl create secret generic woodpecker-secret \
  --namespace <namespace> \
  --from-literal=WOODPECKER_AGENT_SECRET=$(openssl rand -hex 32)

# add credentials for your forge
kubectl create secret generic woodpecker-github-client \
  --namespace <namespace> \
  --from-literal=WOODPECKER_GITHUB_CLIENT=xxxxxxxx

kubectl create secret generic woodpecker-github-secret \
  --namespace <namespace> \
  --from-literal=WOODPECKER_GITHUB_SECRET=xxxxxxxx

# add helm repo
helm repo add woodpecker https://woodpecker-ci.org/
```

### Woodpecker server

```shell
# Install
helm upgrade --install woodpecker-server --namespace <namespace> woodpecker/woodpecker-server

# Uninstall
helm delete woodpecker-server
```

### Woodpecker agent

```shell
# Install
helm upgrade --install woodpecker-agent --namespace <namespace> woodpecker/woodpecker-agent

# Uninstall
helm delete woodpecker-agent
```
