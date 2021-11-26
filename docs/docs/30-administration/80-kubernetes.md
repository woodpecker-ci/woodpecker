# Kubernetes

Woodpecker does not support Kubernetes natively, but being a container first CI engine, it can be deployed to Kubernetes.

## Deploy with HELM

### Preparation

```shell
# create secrets
kubectl create secret generic woodpecker-secret \
  --namespace <namespace> \
  --from-literal=WOODPECKER_AGENT_SECRET=$(openssl rand -hex 32)

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

## Deploy with kubectl

The following yamls represent a server (backed by sqlite and Persistent Volumes) and an agent deployment. The agents can be scaled by the `replica` field.

By design, Woodpecker spins up a new container for each workflow step. It talks to the Docker agent to do that.

However in Kubernetes, the Docker agent is not accessible, therefore this deployment follows a Docker in Docker setup and we deploy a DinD sidecar with the agent.
Build step containers are started up within the agent pod.

Warning: this approach requires `privileged` access. Also DinD's reputation hasn't been too high in the early days of Docker - this changed somewhat over time, and there are organizations succeeding with this approach.

server.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: woodpecker
  namespace: tools
  labels:
    app: woodpecker
spec:
  replicas: 1
  selector:
    matchLabels:
      app: woodpecker
  template:
    metadata:
      labels:
        app: woodpecker
      annotations:
        prometheus.io/scrape: 'true'
    spec:
      containers:
      - image: woodpeckerci/woodpecker-server:latest
        imagePullPolicy: Always
        name: woodpecker
        env:
          - name: "WOODPECKER_ADMIN"
            value: "xxx"
          - name: "WOODPECKER_HOST"
            value: "https://xxx"
          - name: "WOODPECKER_GITHUB"
            value: "true"
          - name: "WOODPECKER_GITHUB_CLIENT"
            value: "xxx"
          - name: "WOODPECKER_GITHUB_SECRET"
            value: "xxx"
          - name: "WOODPECKER_AGENT_SECRET"
            value: "xxx"
        volumeMounts:
          - name: sqlite-volume
            mountPath: /var/lib/woodpecker
      volumes:
        - name: sqlite-volume
          persistentVolumeClaim:
            claimName: woodpecker-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: woodpecker-pvc
  namespace: tools
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: local-path
  resources:
    requests:
      storage: 10Gi
---
kind: Service
apiVersion: v1
metadata:
  name: woodpecker
  namespace: tools
spec:
  type: ClusterIP
  selector:
    app: woodpecker
  ports:
  - protocol: TCP
    name: http
    port: 80
    targetPort: 8000
  - protocol: TCP
    name: grpc
    port: 9000
    targetPort: 9000
---
kind: Ingress
apiVersion: extensions/v1beta1
metadata:
  name: woodpecker
  namespace: tools
spec:
  tls:
  - hosts:
    - xxx
    secretName: xxx
  rules:
  - host: xxx
    http:
      paths:
      - backend:
          serviceName: woodpecker
          servicePort: 80
```

agent.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: woodpecker-agent
  namespace: tools
  labels:
    app: woodpecker-agent
spec:
  selector:
    matchLabels:
      app: woodpecker-agent
  replicas: 2
  template:
    metadata:
      annotations:
      labels:
        app: woodpecker-agent
    spec:
      containers:
      - name: agent
        image: woodpeckerci/woodpecker-agent:latest
        imagePullPolicy: Always
        ports:
        - name: http
          containerPort: 3000
          protocol: TCP
        env:
          - name: WOODPECKER_SERVER
            value: woodpecker.tools.svc.cluster.local:9000
          - name: WOODPECKER_AGENT_SECRET
            value: "xxx"
        resources:
          limits:
            cpu: 2
            memory: 2Gi
        volumeMounts:
        - name: sock-dir
          path: /var/run
      - name: dind
        image: "docker:19.03.5-dind"
        env:
        - name: DOCKER_DRIVER
          value: overlay2
        resources:
          limits:
            cpu: 1
            memory: 2Gi
        securityContext:
          privileged: true
        volumeMounts:
        - name: sock-dir
          mountPath: /var/run
      volumes:
      - name: sock-dir
        emptyDir: {}
```

