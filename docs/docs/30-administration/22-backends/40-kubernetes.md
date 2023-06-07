# Kubernetes backend

:::caution
Kubernetes support is still experimental and not all pipeline features are fully supported yet.

Check the [current state](https://github.com/woodpecker-ci/woodpecker/issues/1513)
:::

The kubernetes backend executes each step inside a newly created pod. A PVC is also created for the lifetime of the pipeline, for transferring files between steps.

## Configuration

### `WOODPECKER_BACKEND_K8S_NAMESPACE`
> Default: `woodpecker`

The namespace to create worker pods in.

### `WOODPECKER_BACKEND_K8S_VOLUME_SIZE`
> Default: `10G`

The volume size of the pipeline volume.

### `WOODPECKER_BACKEND_K8S_STORAGE_CLASS`
> Default: empty

The storage class to use for the pipeline volume.

### `WOODPECKER_BACKEND_K8S_STORAGE_RWX`
> Default: `true`

Determines if `RWX` should be used for the pipeline volume's [access mode](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes). If false, `RWO` is used instead.

### `WOODPECKER_BACKEND_K8S_POD_LABELS`
> Default: empty

Additional labels to apply to worker pods. Must be a YAML object, e.g. `{"example.com/test-label":"test-value"}`.

### `WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS`
> Default: empty

Additional annotations to apply to worker pods. Must be a YAML object, e.g. `{"example.com/test-annotation":"test-value"}`.

## Resources

The kubernetes backend also allows for specifying requests and limits on a per-step basic, most commonly for CPU and memory.

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
        resources:
          requests:
            memory: 128Mi
            cpu: 1000m
          limits:
            memory: 256Mi
```

See the [kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/) for more information on using resources.
