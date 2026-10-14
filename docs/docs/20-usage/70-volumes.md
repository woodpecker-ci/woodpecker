# Volumes

Woodpecker gives the ability to define Docker volumes in the YAML. You can use this parameter to mount files or folders on the host machine into your containers.

:::note
Volumes are only available to trusted repositories and for security reasons should only be used in private environments. See [project settings](./75-project-settings.md#trusted) to enable trusted mode.
:::

```diff
 steps:
   - name: build
     image: docker
     commands:
       - docker build --rm -t octocat/hello-world .
       - docker run --rm octocat/hello-world --test
       - docker push octocat/hello-world
       - docker rmi octocat/hello-world
     volumes:
+      - /var/run/docker.sock:/var/run/docker.sock
```

If you use the Docker backend, you can also use named volumes like `some_volume_name:/var/run/volume`.

Please note that Woodpecker mounts volumes on the host machine. This means you must use absolute paths when you configure volumes. Attempting to use relative paths will result in an error.

```diff
-volumes: [ ./certs:/etc/ssl/certs ]
+volumes: [ /etc/ssl/certs:/etc/ssl/certs ]
```

## Environment variable expansion

Variable references in volume paths are expanded against the environment of the step the volume is used in, e.g. `${CI_WORKSPACE}`:

```diff
 steps:
   - name: install
     image: node:16
     commands:
       - yarn install
     volumes:
+      - /root/node_modules:${CI_WORKSPACE}/node_modules
```

`CI_WORKSPACE` is available in volume paths of every step, so the destination can point into the workspace that the step runs in. Any other variable that is part of the step environment (built-in `CI_*` variables, workflow and step `environment` values) can be used as well. Unknown `${VAR}` references expand to an empty string, mirroring `envsubst` semantics.
