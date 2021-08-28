# Server Setup

## Installation

The below [docker-compose](https://docs.docker.com/compose/) configuration can be used to start Woodpecker with a single agent.

It relies on a number of environment variables that you must set before running `docker-compose up`. The variables are described below.

```yaml
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    ports:
      - 80:8000
      - 9000
    volumes:
      - woodpecker-server-data:/var/lib/drone/
    restart: always
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}

  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
    command: agent
    restart: always
    depends_on:
      - woodpecker-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - WOODPECKER_SERVER=woodpecker-server:9000
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}

volumes:
  woodpecker-server-data:
```

> Each agent is able to process one build by default.
>
> If you have 4 agents installed and connected to the Drone server, your system will process 4 builds in parallel.
>
> You can add more agents to increase the number of parallel builds or set the agent's `WOODPECKER_MAX_PROCS=1` environment variable to increase the number of parallel builds for that agent.


Woodpecker needs to know its own address.

You must therefore provide the address in `<scheme>://<hostname>` format. Please omit trailing slashes.

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
+     - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```

Agents require access to the host machine's Docker daemon.

```diff
services:
  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
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
    image: woodpeckerci/woodpecker-agent:latest
    command: agent
    restart: always
    depends_on: [ woodpecker-server ]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
+     - WOODPECKER_SERVER=woodpecker-server:9000
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```

The server and agents use a shared secret to authenticate communication.

This should be a random string of your choosing and should be kept private. You can generate such string with `openssl rand -hex 32`.

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
+     - WOODPECKER_SECRET=${WOODPECKER_SECRET}
  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
    environment:
      - WOODPECKER_SERVER=woodpecker-server:9000
      - WOODPECKER_DEBUG=true
+     - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```

Registration is closed by default.

This example enables open registration for users that are members of approved GitHub organizations.

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
+     - WOODPECKER_OPEN=true
+     - WOODPECKER_ORGS=dolores,dogpatch
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```

Administrators should also be enumerated in your configuration.

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_ORGS=dolores,dogpatch
+     - WOODPECKER_ADMIN=johnsmith,janedoe
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
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
    image: woodpeckerci/woodpecker-server:latest
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
          - name: "WOODPECKER_SECRET"
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
        image: woodpeckerci/woodpecker-agent:latest
        imagePullPolicy: Always
        ports:
        - name: http
          containerPort: 3000
          protocol: TCP
        env:
          - name: WOODPECKER_SERVER
            value: woodpecker.tools.svc.cluster.local:9000
          - name: WOODPECKER_SECRET
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

## Filtering repositories

Woodpecker operates with the user's OAuth permission. Due to the coarse permission handling of Github, you may end up syncing more repos into Woodpecker than preferred.

Use the `WOODPECKER_REPO_OWNERS` variable to filter which Github user's repos should be synced only. You typically want to put here your company's Github name.

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_ORGS=dolores,dogpatch
+     - WOODPECKER_REPO_OWNERS=mycompany,mycompanyossgithubuser
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```
