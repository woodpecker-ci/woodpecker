# Woodpecker

[Woodpecker](https://woodpecker.laszlo.cloud/) is a fork of the Drone CI system version 0.8, right before the 1.0 release and license changes

## Installing Woodpecker server

### Requirements

```
kubectl create secret generic drone-secret \
  --namespace sre \
  --from-literal=DRONE_SECRET=$(openssl rand -hex 32)
```

[GitHub](https://woodpecker.laszlo.cloud/administration/github/)

```
kubectl create secret generic drone-github-client \
  --namespace <namespace> \
  --from-literal=DRONE_GITHUB_CLIENT=xxxxxxxx
```

```
kubectl create secret generic drone-github-secret \
  --namespace <namespace> \
  --from-literal=DRONE_GITHUB_SECRET=xxxxxxxx
```

```
helm upgrade --install woodpecker-server --namespace <namespace> woodpecker-server/
```


## Installing Woodpecker agent

```
helm upgrade --install woodpecker-agent --namespace <namespace> woodpecker-agent/
```

## Uninstall

```
helm delete woodpecker-agent
helm delete woodpecker-server
```

## Support

For questions, suggestions, and discussion, visit the [Discord](https://discord.gg/fcMQqSMXJy).
