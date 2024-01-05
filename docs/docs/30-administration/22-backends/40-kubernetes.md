# Kubernetes backend

The kubernetes backend executes steps inside standalone pods. A temporary PVC is created for the lifetime of the pipeline to transfer files between steps.

## General Configuration

These env vars can be set in the `env:` sections of both `server` and `agent`.
They do not need to be set for both but only for the part to which it is relevant to.

```yaml
server:
  env:
    WOODPECKER_SESSION_EXPIRES: "300h"
    [...]

agent:
  env:
    [...]
```

### Options

#### `WOODPECKER_BACKEND_K8S_NAMESPACE`

> Default: `woodpecker`

The namespace to create worker pods in.

#### `WOODPECKER_BACKEND_K8S_VOLUME_SIZE`

> Default: `10G`

The volume size of the pipeline volume.

#### `WOODPECKER_BACKEND_K8S_STORAGE_CLASS`

> Default: empty

The storage class to use for the pipeline volume.

#### `WOODPECKER_BACKEND_K8S_STORAGE_RWX`

> Default: `true`

Determines if `RWX` should be used for the pipeline volume's [access mode](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes). If false, `RWO` is used instead.

#### `WOODPECKER_BACKEND_K8S_POD_LABELS`

> Default: empty

Additional labels to apply to worker pods. Must be a YAML object, e.g. `{"example.com/test-label":"test-value"}`.

#### `WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS`

> Default: empty

Additional annotations to apply to worker pods. Must be a YAML object, e.g. `{"example.com/test-annotation":"test-value"}`.

#### `WOODPECKER_BACKEND_K8S_SECCTX_NONROOT`

> Default: `false`

Determines if containers must be required to run as non-root users.

## Job specific configuration

### Resources

The kubernetes backend also allows for specifying requests and limits on a per-step basic, most commonly for CPU and memory.
We recommend to add a `resources` definition to all steps to ensure efficient scheduling.

Here is an example definition with an arbitrary `resources` definition below the `backend_options` section:

```yaml
steps:
  'My kubernetes step':
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

### serviceAccountName

Specify the name of the ServiceAccount which the build pod will mount. This serviceAccount must be created externally.
See the [kubernetes documentation](https://kubernetes.io/docs/concepts/security/service-accounts/) for more information on using serviceAccounts.

### nodeSelector

Specifies the label which is used to select the node on which the job will be executed.

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

### tolerations

When you use nodeSelector and the node pool is configured with Taints, you need to specify the Tolerations. Tolerations allow the scheduler to schedule pods with matching taints.
See the [kubernetes documentation](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) for more information on using tolerations.

Example pipeline configuration:

```yaml
steps:
  build:
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
          beta.kubernetes.io/instance-type: p3.8xlarge
        tolerations:
          - key: 'key1'
            operator: 'Equal'
            value: 'value1'
            effect: 'NoSchedule'
            tolerationSeconds: 3600
```

### Volumes

To mount volumes a persistent volume (PV) and persistent volume claim (PVC) are needed on the cluster which can be referenced in steps via the `volume:` option.
Assuming a PVC named "woodpecker-cache" exists, it can be referenced as follows in a step:

```yaml
steps:
  "Restore Cache":
    image: meltwater/drone-cache
    volumes:
      - woodpecker-cache:/woodpecker/src/cache
    settings:
      mount:
        - "woodpecker-cache"
    [...]
```

### `securityContext`

Use the following configuration to set the `securityContext` for the pod/container running a given pipeline step:

```yaml
steps:
  test:
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

Note that the `backend_options.kubernetes.securityContext` object allows you to set both pod and container level security context options in one object.
By default, the properties will be set at the pod level. Properties that are only supported on the container level will be set there instead. So, the
configuration shown above will result in something like the following pod spec:

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

See the [kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) for more information on using `securityContext`.

## Tips and tricks

### CRI-O

CRI-O users currently need to configure the workspace for all workflows in order for them to run correctly. Add the following at the beginning of your configuration:

```yaml
workspace:
  base: '/woodpecker'
  path: '/'
```

See [this issue](https://github.com/woodpecker-ci/woodpecker/issues/2510) for more details.
