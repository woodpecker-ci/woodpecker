# Services

Woodpecker provides a services section in the YAML file used for defining service containers.
The below configuration composes database and cache containers.

Services are accessed using custom hostnames.
In the example below, the MySQL service is assigned the hostname `database` and is available at `database:3306`.

```yaml
steps:
  - name: build
    image: golang
    commands:
      - go build
      - go test

services:
  - name: database
    image: mysql

  - name: cache
    image: redis
```

You can define a port and a protocol explicitly:

```yaml
services:
  - name: database
    image: mysql
    ports:
      - 3306

  - name: wireguard
    image: wg
    ports:
      - 51820/udp
```

## Configuration

Service containers generally expose environment variables to customize service startup such as default usernames, passwords and ports. Please see the official image documentation to learn more.

```diff
 services:
   - name: database
     image: mysql
+    environment:
+      - MYSQL_DATABASE=test
+      - MYSQL_ALLOW_EMPTY_PASSWORD=yes

   - name: cache
     image: redis
```

## Detachment

Service and long running containers can also be included in the pipeline section of the configuration using the detach parameter without blocking other steps. This should be used when explicit control over startup order is required.

```diff
 steps:
   - name: build
     image: golang
     commands:
       - go build
       - go test

   - name: database
     image: redis
+    detach: true

   - name: test
     image: golang
     commands:
       - go test
```

Containers from detached steps will terminate when the pipeline ends.

## Initialization

Service containers require time to initialize and begin to accept connections. If you are unable to connect to a service you may need to wait a few seconds or implement a backoff.

```diff
 steps:
   - name: test
     image: golang
     commands:
+      - sleep 15
       - go get
       - go test

 services:
   - name: database
     image: mysql
```

## Complete Pipeline Example

```yaml
services:
  - name: database
    image: mysql
    environment:
      - MYSQL_DATABASE=test
      - MYSQL_ROOT_PASSWORD=example
steps:
  - name: get-version
    image: ubuntu
    commands:
      - ( apt update && apt dist-upgrade -y && apt install -y mysql-client 2>&1 )> /dev/null
      - sleep 30s # need to wait for mysql-server init
      - echo 'SHOW VARIABLES LIKE "version"' | mysql -uroot -hdatabase test -pexample
```
