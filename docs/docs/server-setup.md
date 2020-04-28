## Installation

The below [docker-compose](https://docs.docker.com/compose/) configuration can be used to start Woodpecker with a single agent.

It relies on a number of environment variables that you must set before running `docker-compose up`. The variables are described below.

```yaml
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    ports:
      - 80:8000
      - 9000
    volumes:
      - woodpecker-server-data:/var/lib/drone/
    restart: always
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}

  woodpecker-agent:
    image: laszlocloud/woodpecker-agent:v0.9.0
    command: agent
    restart: always
    depends_on:
      - woodpecker-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_SERVER=woodpecker-server:9000
      - DRONE_SECRET=${DRONE_SECRET}

volumes:
  woodpecker-server-data:
```

> Each agent is able to process one build by default.
>
> If you have 4 agents installed and connected to the Drone server, your system will process 4 builds in parallel.
>
> You can add more agents to increase the number of parallel builds or set the agent's `DRONE_MAX_PROCS=1` environment variable to increase the number of parallel builds for that agent.


Woodpecker needs to know its own address.

You must therefore provide the address in `<scheme>://<hostname>` format. Please omit trailing slashes.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
      - DRONE_OPEN=true
+     - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
```

Agents require access to the host machine's Docker daemon.

```diff
services:
  woodpecker-agent:
    image: laszlocloud/woodpecker-agent:v0.9.0
    command: agent
    restart: always
    depends_on: [ woodpecker-server ]
+   volumes:
+     - /var/run/docker.sock:/var/run/docker.sock
```

Agents require the server address for agent-to-server communication.

```diff
services:
  woodpecker-agent:
    image: laszlocloud/woodpecker-agent:v0.9.0
    command: agent
    restart: always
    depends_on: [ woodpecker-server ]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
+     - DRONE_SERVER=woodpecker-server:9000
      - DRONE_SECRET=${DRONE_SECRET}
```

The server and agents use a shared secret to authenticate communication.

This should be a random string of your choosing and should be kept private. You can generate such string with `openssl rand -hex 32`.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
+     - DRONE_SECRET=${DRONE_SECRET}
  woodpecker-agent:
    image: laszlocloud/woodpecker-agent:v0.9.0
    environment:
      - DRONE_SERVER=woodpecker-server:9000
      - DRONE_DEBUG=true
+     - DRONE_SECRET=${DRONE_SECRET}
```

Registration is closed by default.

This example enables open registration for users that are members of approved GitHub organizations.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
+     - DRONE_OPEN=true
+     - DRONE_ORGS=dolores,dogpatch
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
```

Administrators should also be enumerated in your configuration.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
      - DRONE_OPEN=true
      - DRONE_ORGS=dolores,dogpatch
+     - DRONE_ADMIN=johnsmith,janedoe
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
```


## Authentication

Authentication is done using OAuth and is delegated to one of multiple version control providers, configured using environment variables. The example above demonstrates basic GitHub integration.

See the complete reference for [Github](/administration/github), [Bitbucket Cloud](/administration/bitbucket), [Bitbucket Server](/administration/bitbucket_server) and [Gitlab](/administration/gitlab).

## Database

Woodpecker mounts a [data volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to persist the sqlite database.

See the [database settings](/administration/database) page to configure Postgresql or MySQL as database.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    ports:
      - 80:8000
      - 9000
+   volumes:
+     - woodpecker-server-data:/var/lib/drone/
    restart: always
```

## SSL

Woodpecker supports ssl configuration by mounting certificates into your container. See the [SSL guide](/administration/ssl).

Automated [Lets Encrypt](/administration/lets-encrypt) is also supported.

## Metrics

A [Prometheus endpoint](/administration/prometheus) is exposed.

## Behind a proxy

See the [proxy guide](/administration/proxy) if you want to see a setup behind Apache, Nginx, Caddy or ngrok.

## Deploying on Kubernetes

Woodpecker does not support Kubernetes natively, but being a container first CI engine, it can be deployed to Kubernetes.

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
      - image: laszlocloud/woodpecker-server:v0.9.2
        imagePullPolicy: Always
        name: woodpecker
        env:
          - name: "DRONE_ADMIN"
            value: "xxx"
          - name: "DRONE_HOST"
            value: "https://xxx"
          - name: "DRONE_GITHUB"
            value: "true"
          - name: "DRONE_GITHUB_CLIENT"
            value: "xxx"
          - name: "DRONE_GITHUB_SECRET"
            value: "xxx"
          - name: "DRONE_SECRET"
            value: "xxx"
        volumeMounts:
          - name: sqlite-volume
            mountPath: /var/lib/drone
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
        image: laszlocloud/woodpecker-agent:v0.9.2
        imagePullPolicy: Always
        ports:
        - name: http
          containerPort: 3000
          protocol: TCP
        env:
          - name: DRONE_SERVER
            value: woodpecker.tools.svc.cluster.local:9000
          - name: DRONE_SECRET
            value: "xxx"
          - name: DOCKER_HOST
            value: tcp://localhost:2375
        resources:
          limits:
            cpu: 2
            memory: 2Gi
      - name: dind
        image: "docker:19.03.5-dind"
        env:
        - name: DOCKER_DRIVER
          value: overlay2
        - name: DOCKER_TLS_CERTDIR
          value: "" # due to https://github.com/docker-library/docker/pull/166 & https://gitlab.com/gitlab-org/gitlab-runner/issues/4512
        resources:
          limits:
            cpu: 1
            memory: 2Gi
        securityContext:
          privileged: true
```
