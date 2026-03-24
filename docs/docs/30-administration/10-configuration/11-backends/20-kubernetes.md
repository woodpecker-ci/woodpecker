---
toc_max_heading_level: 2
---

# Kubernetes

The Kubernetes backend executes steps inside standalone Pods. A temporary PVC is created for the lifetime of the pipeline to transfer files between steps.

## Metadata labels

Woodpecker adds some labels to the pods to provide additional context to the workflow. These labels can be used for various purposes, e.g. for simple debugging or as selectors for network policies.

The following metadata labels are supported:

- `woodpecker-ci.org/forge-id`
- `woodpecker-ci.org/repo-forge-id`
- `woodpecker-ci.org/repo-id`
- `woodpecker-ci.org/repo-name`
- `woodpecker-ci.org/repo-full-name`
- `woodpecker-ci.org/branch`
- `woodpecker-ci.org/org-id`
- `woodpecker-ci.org/task-uuid`
- `woodpecker-ci.org/step`

## Private registries

In addition to [registries specified in the UI](../../../20-usage/41-registries.md), you may provide [registry credentials in Kubernetes Secrets](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/) to pull private container images defined in your pipeline YAML.

Place these Secrets in namespace defined by `WOODPECKER_BACKEND_K8S_NAMESPACE` and provide the Secret names to Agents via `WOODPECKER_BACKEND_K8S_PULL_SECRET_NAMES`.

## Step specific configuration

### Resources

The Kubernetes backend also allows for specifying requests and limits on a per-step basic, most commonly for CPU and memory.
We recommend to add a `resources` definition to all steps to ensure efficient scheduling.

Here is an example definition with an arbitrary `resources` definition below the `backend_options` section:

```yaml
steps:
  - name: 'My kubernetes step'
    image: alpine
    commands:
      - echo "Hello world"
    backend_options:
      kubernetes:
        resources:
          requests:
            memory: 200Mi
            cpu: 100m
          limits:
            memory: 400Mi
            cpu: 1000m
```

You can use [Limit Ranges](https://kubernetes.io/docs/concepts/policy/limit-range/) if you want to set the limits by per-namespace basis.

### Runtime class

`runtimeClassName` specifies the name of the RuntimeClass which will be used to run this Pod. If no `runtimeClassName` is specified, the default RuntimeHandler will be used.
See the [Kubernetes documentation](https://kubernetes.io/docs/concepts/containers/runtime-class/) for more information on specifying runtime classes.

### Service account

`serviceAccountName` specifies the name of the ServiceAccount which the Pod will mount. This service account must be created externally.
See the [Kubernetes documentation](https://kubernetes.io/docs/concepts/security/service-accounts/) for more information on using service accounts.

```yaml
steps:
  - name: 'My kubernetes step'
    image: alpine
    commands:
      - echo "Hello world"
    backend_options:
      kubernetes:
        # Use the service account `default` in the current namespace.
        # This usually the same as wherever woodpecker is deployed.
        serviceAccountName: default
```

To give steps access to the Kubernetes API via service account, take a look at [RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)

### Node selector

`nodeSelector` specifies the labels which are used to select the node on which the step will be executed.

Labels defined here will be appended to a list which already contains `"kubernetes.io/arch"`.
By default `"kubernetes.io/arch"` is inferred from the agents' platform. One can override it by setting that label in the `nodeSelector` section of the `backend_options`.
Without a manual overwrite, builds will be randomly assigned to the runners and inherit their respective architectures.

To overwrite this, one needs to set the label in the `nodeSelector` section of the `backend_options`.
A practical example for this is when running a matrix-build and delegating specific elements of the matrix to run on a specific architecture.
In this case, one must define an arbitrary key in the matrix section of the respective matrix element:

```yaml
matrix:
  include:
    - NAME: runner1
      ARCH: arm64
```

And then overwrite the `nodeSelector` in the `backend_options` section of the step(s) using the name of the respective env var:

```yaml
[...]
    backend_options:
      kubernetes:
        nodeSelector:
          kubernetes.io/arch: "${ARCH}"
```

You can use [WOODPECKER_BACKEND_K8S_POD_NODE_SELECTOR](#backend_k8s_pod_node_selector) if you want to set the node selector per Agent
or [PodNodeSelector](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podnodeselector) admission controller if you want to set the node selector by per-namespace basis.

### Tolerations

When you use `nodeSelector` and the node pool is configured with Taints, you need to specify the Tolerations. Tolerations allow the scheduler to schedule Pods with matching taints.
See the [Kubernetes documentation](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) for more information on using tolerations.

Example pipeline configuration:

```yaml
steps:
  - name: build
    image: golang
    commands:
      - go get
      - go build
      - go test
    backend_options:
      kubernetes:
        serviceAccountName: 'my-service-account'
        resources:
          requests:
            memory: 128Mi
            cpu: 1000m
          limits:
            memory: 256Mi
        nodeSelector:
          beta.kubernetes.io/instance-type: Standard_D2_v3
        tolerations:
          - key: 'key1'
            operator: 'Equal'
            value: 'value1'
            effect: 'NoSchedule'
            tolerationSeconds: 3600
        affinity:
          nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              nodeSelectorTerms:
                - matchExpressions:
                    - key: topology.kubernetes.io/zone
                      operator: In
                      values:
                        - eu-central-1a
                        - eu-central-1b
```

### Affinity

Kubernetes [affinity and anti-affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) rules allow you to constrain which nodes your pods can be scheduled on based on node labels, or co-locate/spread pods relative to other pods.

You can configure affinity at two levels:

1. **Per-step via `backend_options.kubernetes.affinity`** (shown in example above) - requires agent configuration to allow it
2. **Agent-wide via `WOODPECKER_BACKEND_K8S_POD_AFFINITY`** - applies to all pods unless overridden

#### Agent-wide affinity

To apply affinity rules to all workflow pods, configure the agent with YAML-formatted affinity:

```yaml
WOODPECKER_BACKEND_K8S_POD_AFFINITY: |
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: node-role.kubernetes.io/worker
              operator: In
              values:
                - "true"
```

By default, per-step affinity settings are **not allowed** for security reasons. To enable them:

```bash
WOODPECKER_BACKEND_K8S_POD_AFFINITY_ALLOW_FROM_STEP: true
```

:::warning
Enabling `WOODPECKER_BACKEND_K8S_POD_AFFINITY_ALLOW_FROM_STEP` in multi-tenant environments allows pipeline authors to control pod placement, which may have security or resource isolation implications.
:::

When per-step affinity is allowed and specified, it **replaces** the agent-wide affinity entirely (not merged).

#### Example: agent affinity for co-location

This example configures all workflow pods within a workflow to be co-located on the same node, while requiring other workflows run on different nodes.

It uses `matchLabelKeys` to dynamically match pods with the same `woodpecker-ci.org/task-uuid`, and `mismatchLabelKeys` to separating pods with different task UUIDs:

```yaml
WOODPECKER_BACKEND_K8S_POD_AFFINITY: |
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector: {}
      matchLabelKeys:
        - woodpecker-ci.org/task-uuid
      topologyKey: "kubernetes.io/hostname"
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector: {}
      mismatchLabelKeys:
      - woodpecker-ci.org/task-uuid
      topologyKey: "kubernetes.io/hostname"
```

:::note
The `matchLabelKeys` and `mismatchLabelKeys` features require Kubernetes v1.29+ (alpha with feature gate `MatchLabelKeysInPodAffinity`) or v1.33+ (beta, enabled by default). These fields allow the Kubernetes API server to dynamically populate label selectors at pod creation time, eliminating the need to hardcode values like `$(WOODPECKER_TASK_UUID)`.
:::

#### Example: Node affinity for GPU workloads

Ensure a step runs only on GPU-enabled nodes:

```yaml
steps:
  - name: train-model
    image: tensorflow/tensorflow:latest-gpu
    backend_options:
      kubernetes:
        affinity:
          nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              nodeSelectorTerms:
                - matchExpressions:
                    - key: accelerator
                      operator: In
                      values:
                        - nvidia-tesla-v100
```

### Volumes

To mount volumes a PersistentVolume (PV) and PersistentVolumeClaim (PVC) are needed on the cluster which can be referenced in steps via the `volumes` option.

Persistent volumes must be created manually. Use the Kubernetes [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) documentation as a reference.

_If your PVC is not highly available or NFS-based, use the `affinity` settings (documented above) to ensure that your steps are executed on the correct node._

NOTE: If you plan to use this volume in more than one workflow concurrently, make sure you have configured the PVC in `RWX` mode. Keep in mind that this feature must be supported by the used CSI driver:

```yaml
accessModes:
  - ReadWriteMany
```

Assuming a PVC named `woodpecker-cache` exists, it can be referenced as follows in a plugin step:

```yaml
steps:
  - name: "Restore Cache"
    image: meltwater/drone-cache
    volumes:
      - woodpecker-cache:/woodpecker/src/cache
    settings:
      mount:
        - "woodpecker-cache"
    [...]
```

Or as follows when using a normal image:

```yaml
steps:
  - name: "Edit cache"
    image: alpine:latest
    volumes:
      - woodpecker-cache:/woodpecker/src/cache
    commands:
      - echo "Hello World" > /woodpecker/src/cache/output.txt
    [...]
```

### Security context

Use the following configuration to set the [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) for the Pod/container running a given pipeline step:

```yaml
steps:
  - name: test
    image: alpine
    commands:
      - echo Hello world
    backend_options:
      kubernetes:
        securityContext:
          runAsUser: 999
          runAsGroup: 999
          privileged: true
    [...]
```

Note that the `backend_options.kubernetes.securityContext` object allows you to set both Pod and container level security context options in one object.
By default, the properties will be set at the Pod level. Properties that are only supported on the container level will be set there instead. So, the
configuration shown above will result in something like the following Pod spec:

<!-- cspell:disable -->

```yaml
kind: Pod
spec:
  securityContext:
    runAsUser: 999
    runAsGroup: 999
  containers:
    - name: wp-01hcd83q7be5ymh89k5accn3k6-0-step-0
      image: alpine
      securityContext:
        privileged: true
  [...]
```

<!-- cspell:enable -->

You can also restrict a syscalls of containers with [seccomp](https://kubernetes.io/docs/tutorials/security/seccomp/) profile.

```yaml
backend_options:
  kubernetes:
    securityContext:
      seccompProfile:
        type: Localhost
        localhostProfile: profiles/audit.json
```

or restrict a container's access to resources by specifying [AppArmor](https://kubernetes.io/docs/tutorials/security/apparmor/) profile

```yaml
backend_options:
  kubernetes:
    securityContext:
      apparmorProfile:
        type: Localhost
        localhostProfile: k8s-apparmor-example-deny-write
```

or configure a specific `fsGroupChangePolicy` (Kubernetes defaults to 'Always')

```yaml
backend_options:
  kubernetes:
    securityContext:
      fsGroupChangePolicy: OnRootMismatch
```

:::note
The feature requires Kubernetes v1.30 or above.
:::

You can set `allowPrivilegeEscalation` to `false` to prevent a container from gaining more privileges than its parent process.

```yaml
backend_options:
  kubernetes:
    securityContext:
      allowPrivilegeEscalation: false
```

You can also drop [Linux capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html) from a container. Adding capabilities is not allowed.

```yaml
backend_options:
  kubernetes:
    securityContext:
      capabilities:
        drop:
          - ALL
```

### Annotations and labels

You can specify arbitrary [annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) and [labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) to be set on the Pod definition for a given workflow step using the following configuration:

```yaml
backend_options:
  kubernetes:
    annotations:
      workflow-group: alpha
      io.kubernetes.cri-o.Devices: /dev/fuse
    labels:
      environment: ci
      app.kubernetes.io/name: builder
```

In order to enable this configuration you need to set the appropriate environment variables to `true` on the woodpecker agent:
[WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS_ALLOW_FROM_STEP](#backend_k8s_pod_annotations_allow_from_step) and/or [WOODPECKER_BACKEND_K8S_POD_LABELS_ALLOW_FROM_STEP](#backend_k8s_pod_labels_allow_from_step).

## Tips and tricks

### CRI-O

CRI-O users currently need to configure the workspace for all workflows in order for them to run correctly. Add the following at the beginning of your configuration:

```yaml
workspace:
  base: '/woodpecker'
  path: '/'
```

See [this issue](https://github.com/woodpecker-ci/woodpecker/issues/2510) for more details.

### `KUBERNETES_SERVICE_HOST` environment variable

Like the below env vars used for configuration, this can be set in the environment for configuration of the agent.
It configures the address of the Kubernetes API server to connect to.

If running the agent within Kubernetes, this will already be set and you don't have to add it manually.

### Headless services

For each workflow run a [headless services](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) is created,
and all steps asigned the subdomain that matches the headless service, so any step can reach other steps via DNS by using the step name as hostname.

Using the headless services, the step pod is connected to directly, so any port on the other step pods can be reached.

This is useful for some use-cases, like test-containers in a docker-in-docker setup, where the step needs to connect to many ports on the docker host service.

```yaml
steps:
  - name: test
    image: docker:cli # use 'docker:<major-version>-cli' or similar in production
    environment:
      DOCKER_HOST: 'tcp://docker:2376'
      DOCKER_CERT_PATH: '/woodpecker/dind-certs/client'
      DOCKER_TLS_VERIFY: '1'
    commands:
      - docker run hello-world

  - name: docker
    image: docker:dind # use 'docker:<major-version>-dind' or similar in production
    detached: true
    privileged: true
    environment:
      DOCKER_TLS_CERTDIR: /woodpecker/dind-certs
```

If ports are defined on a service, then woodpecker will create a normal service for the pod, which use hosts override using the services cluster IP.

## Running in an unprivileged namespace

Woodpecker by default requires the namespace where workflow pods run to be privileged.

However, it's possible to configure the agent in such a way that allows workflow pods to run in an unprivileged namespace.
This comes with some drawbacks and it's the reason why its disabled by default.

Major drawbacks are:

- You won't be able to use commands like `apk add` or `apt install` in most images.
  The easiest way to workaround this is by building your own image with the tools you require already prebundled in it.
  This also have the advantage that workflows will run faster, since it won't need to fetch packages during each run.
- If you need to build Docker/OCI images, you'll need to use a rootless builder like Buildah or BuildKit in rootless mode.
- The default clone step currently doesn't work in unprivileged namespaces, however its possible to define your own clone step that can run unprivileged. More details below.

Please note, this guide assumes you already have a working woodpecker instance running in your kubernetes cluster in a privileged namespace.

### Setting security context environment variables

Depending on how you installed the woodpecker server and agent, this step may be different.
To make this guide as generic as possible, we will only list the environment variables that need to be updated.

On your woodpecker-agent Deployment/StatefulSet, set this environment variables:

```sh
WOODPECKER_BACKEND_K8S_DEFAULT_SECCTX='{"runAsUser":1000,"runAsGroup":1000,"fsGroup":1000,"fsGroupChangePolicy": "OnRootMismatch"}'
WOODPECKER_BACKEND_K8S_ENFORCED_SECCTX='{"privileged":false,"runAsNonRoot":true,"allowPrivilegeEscalation":false,"seccompProfile": {"type": "RuntimeDefault"}, "capabilities": {"drop": ["ALL"]}}'
```

Wait until the update rolls out.

### Setting up the namespace

Make the namespace where woodpecker worker pods run restricted, if you haven't done it yet:

```sh
kubectl label namespace woodpecker \
  pod-security.kubernetes.io/enforce=restricted \
  pod-security.kubernetes.io/audit=restricted \
  pod-security.kubernetes.io/warn=restricted \
  --overwrite
```

Please note here we use the namespace name `woodpecker`, but you should replace it with the actual namespace name you're using for woodpecker worker pods. If you have set `WOODPECKER_BACKEND_K8S_NAMESPACE`, then this is the namespace you should update. If you haven't, worker pods will run by default in the same namespace as the `woodpecker-agent`.

### Unprivileged clone step

Currently, the default git clone step depends on the kubernetes container runtime to create its working directory.
Most container runtimes will create it owned by root by default, which will make the plugin fail with `Permission denied` errors if we dont precreate it, since the container will run unprivileged.

Also, the default git clone plugin will use /app as its home, which is owned by root and writable only by root in the image, so we'll need to change that too.

This is how our workflow should look like:

```yaml
# skip the default clone step since we're replacing it with our own.
skip_clone: true

steps:
  # precreate the `plugin-git` working directory, so it won't fail with `Permission denied` errors later.
  - name: prepare
    image: alpine
    commands:
      - mkdir -p $CI_WORKSPACE

  - name: clone
    image: quay.io/woodpeckerci/plugin-git
    settings:
      # set home to /tmp, which is writable by everybody in the `plugin-git` image.
      home: /tmp
```

### Final notes about unprivileged namespaces

Please note this setup is experimental, and you may encounter permission issues with other plugins.

## Environment variables

These env vars can be set in the `env:` sections of the agent.

---

### BACKEND_K8S_NAMESPACE

- Name: `WOODPECKER_BACKEND_K8S_NAMESPACE`
- Default: `woodpecker`

The namespace to create worker Pods in.

---

### BACKEND_K8S_NAMESPACE_PER_ORGANIZATION

- Name: `WOODPECKER_BACKEND_K8S_NAMESPACE_PER_ORGANIZATION`
- Default: `false`

Enables namespace isolation per Woodpecker organization. When enabled, each organization gets its own dedicated Kubernetes namespace for improved security and resource isolation.

With this feature enabled, Woodpecker creates separate Kubernetes namespaces for each organization using the format `{WOODPECKER_BACKEND_K8S_NAMESPACE}-{organization-id}`. Namespaces are created automatically when needed, but they are not automatically deleted when organizations are removed from Woodpecker.

### BACKEND_K8S_VOLUME_SIZE

- Name: `WOODPECKER_BACKEND_K8S_VOLUME_SIZE`
- Default: `10G`

The volume size of the pipeline volume.

---

### BACKEND_K8S_STORAGE_CLASS

- Name: `WOODPECKER_BACKEND_K8S_STORAGE_CLASS`
- Default: none

The storage class to use for the pipeline volume.

---

### BACKEND_K8S_STORAGE_RWX

- Name: `WOODPECKER_BACKEND_K8S_STORAGE_RWX`
- Default: `true`

Determines if `RWX` should be used for the pipeline volume's [access mode](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes). If false, `RWO` is used instead.

---

### BACKEND_K8S_POD_LABELS

- Name: `WOODPECKER_BACKEND_K8S_POD_LABELS`
- Default: none

Additional labels to apply to worker Pods. Must be a YAML object, e.g. `{"example.com/test-label":"test-value"}`.

---

### BACKEND_K8S_POD_LABELS_ALLOW_FROM_STEP

- Name: `WOODPECKER_BACKEND_K8S_POD_LABELS_ALLOW_FROM_STEP`
- Default: `false`

Determines if additional Pod labels can be defined from a step's backend options.

---

### BACKEND_K8S_POD_ANNOTATIONS

- Name: `WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS`
- Default: none

Additional annotations to apply to worker Pods. Must be a YAML object, e.g. `{"example.com/test-annotation":"test-value"}`.

---

### BACKEND_K8S_POD_ANNOTATIONS_ALLOW_FROM_STEP

- Name: `WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS_ALLOW_FROM_STEP`
- Default: `false`

Determines if Pod annotations can be defined from a step's backend options.

---

### BACKEND_K8S_POD_TOLERATIONS

- Name: `WOODPECKER_BACKEND_K8S_POD_TOLERATIONS`
- Default: none

Additional tolerations to apply to worker Pods. Must be a YAML object, e.g. `[{"effect":"NoSchedule","key":"jobs","operator":"Exists"}]`.

---

### BACKEND_K8S_POD_TOLERATIONS_ALLOW_FROM_STEP

- Name: `WOODPECKER_BACKEND_K8S_POD_TOLERATIONS_ALLOW_FROM_STEP`
- Default: `true`

Determines if Pod tolerations can be defined from a step's backend options.

---

### BACKEND_K8S_POD_NODE_SELECTOR

- Name: `WOODPECKER_BACKEND_K8S_POD_NODE_SELECTOR`
- Default: none

Additional node selector to apply to worker pods. Must be a YAML object, e.g. `{"topology.kubernetes.io/region":"eu-central-1"}`.

---

### BACKEND_K8S_SECCTX_NONROOT <!-- cspell:ignore SECCTX NONROOT -->

- Name: `WOODPECKER_BACKEND_K8S_SECCTX_NONROOT`
- Default: `false`

Determines if containers must be required to run as non-root users.

---

### BACKEND_K8S_DEFAULT_SECCTX <!-- cspell:ignore SECCTX NONROOT -->

- Name: `WOODPECKER_BACKEND_K8S_DEFAULT_SECCTX`
- Default: none

The default security context that will be applied to all step pods.

Must be a YAML object, e.g. `{"runAsUser":1000,"runAsGroup":1000,"fsGroup":1000,"fsGroupChangePolicy": "OnRootMismatch"}`

The security context defined here can be overriden by workflow steps. If you want to define a security context that cannot be overriden, check the next option.

---

### BACKEND_K8S_ENFORCED_SECCTX <!-- cspell:ignore SECCTX NONROOT -->

- Name: `WOODPECKER_BACKEND_K8S_ENFORCED_SECCTX`
- Default: none

The security context that will be applied to all step pods. Cannot be overriden by workflow steps.

Must be a YAML object, e.g. `{"privileged":false,"runAsNonRoot":true,"allowPrivilegeEscalation":false,"seccompProfile": {"type": "RuntimeDefault"}, "capabilities": {"drop": ["ALL"]}}`

---

### BACKEND_K8S_PULL_SECRET_NAMES

- Name: `WOODPECKER_BACKEND_K8S_PULL_SECRET_NAMES`
- Default: none

Secret names to pull images from private repositories. See, how to [Pull an Image from a Private Registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

---

### BACKEND_K8S_PRIORITY_CLASS

- Name: `WOODPECKER_BACKEND_K8S_PRIORITY_CLASS`
- Default: none, which will use the default priority class configured in Kubernetes

Which [Kubernetes PriorityClass](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/priority-class-v1/) to assign to created job pods.
