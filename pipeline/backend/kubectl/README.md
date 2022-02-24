# Kubernetes backend using `kubectl`

This backend executes tasks on kubernetes managed pods. See [environment variables](#envs) for configuration.

Backend name: `kubectl`

### Requirements

1. The executable `kubectl` must on both the agent and the server.

### Supports

1. Environment variables for configuration.
1. Network isolation and internet access.
1. Configuration control through [environment variables](#envs).
1. Container features:
   ```yaml
   commands: [string],
   image: string,
   environment: {},
   pull: bool,
   detached: bool,
   privileged: bool,
   alias: string,
   dns: [string],
   dnssearch: [string],
   ```

### Dose not support

1. Local volumes.
1. Multiple volumes.
1. Volumes on detached steps (or services), without a special PVC class
1. Configuration control through CLI commands.

### Mode of operation

The backend creates, for each pipeline run, the following resources,

1. A persistent volume claim (PVC), that is mounted on individual steps.
1. A network policy for the run (if enabled)

Once created, for each step (or clone, or service), the backend creates a [Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) (apiVersion:batch/v1), which executes the step. Where,

1. Detached jobs are executed **without** mounting the PVC, and are **always** deleted at the end of the run.
1. Steps are executed after the PVC is loaded.

Once finished (failed or succeeded), the resources are deleted according to the delete policy, which can be one of,

1.  IfFailed
1.  IfSucceeded
1.  Always
1.  Never

# Networking isolation (with internet access)

**ONLY applies when**: `WOODPECKER_KUBECTL_ENABLE_NETWORK_POLICY=true`

<a name="network-isolation"></a>

To allow networking isolation, with internet access, there **must** exist in the **same** namespace where the pods are executed a global network policy that would act on all woodpecker pods. If the policy dose not exist, the pod would **only be allowed to access** pods which are part of the same pipeline.

Example policy:

```yaml
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  # namespace: [your cicd namespace]
  name: woodpecker-jobs
spec:
  podSelector:
    matchLabels:
      # Match all woodpecker pods.
      woodpecker: "true"
  policyTypes:
    - Egress
  egress:
    # Allow specific services and pods.
    - to:
        # Allows dns resolution.
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              k8s-app: kube-dns
          # Allows internet acces, except for internal pods.
        - ipBlock:
            cidr: 0.0.0.0/0 # Allow all ips
            # Your services and pods CIDR, that would be excluded.
            except:
              - 172.0.0.0/8 # All pods in the cluster
              - 173.0.0.0/8 # All services (svc) in the cluster.
              # anything else?
```

## Environment variables

<a name="envs"></a>

Kubectl/Kubernetes:

1. `WOODPECKER_KUBECTL_EXECUTABLE` The path to the kubectl executable. Defaults to `kubectl`
1. `WOODPECKER_KUBECTL_NAMESPACE` The namespace to execute in. Auto detected if empty.
1. `WOODPECKER_KUBECTL_CONTEXT` The context to execute in. Auto detected if empty.

Job:

1. `WOODPECKER_KUBECTL_MEMORY_LIMIT` The memory limit in `Mi/Gi`. e.g. `200Mi`
1. `WOODPECKER_KUBECTL_CPU_LIMIT` The cpu limit in cpu counts. `m` = mili. e.g. `200m`, `1`
1. `WOODPECKER_KUBECTL_DELETE_POLICY` The delete policy. One of `IfFailed/IfSucceeded/Always/Never`
1. `WOODPECKER_KUBECTL_FORCE_PULL_POLICY` If exists. Will ignore the `pull` yaml value. See [policies](https://kubernetes.io/docs/concepts/containers/images/).
1. `WOODPECKER_KUBECTL_TERMINATION_GRACE_PERIOD` The number of seconds before a job is forcefully terminated (SIGKILL).
1. `WOODPECKER_KUBECTL_ENABLE_NETWORK_POLICY` If true, will deploy a network policy for the run. See [Network Isolation](#network-isolation)

Persistent volumes:

1. `WOODPECKER_KUBECTL_PVC_STORAGE_SIZE` The volume size, by default `1Gi`. See [PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/).
1. `WOODPECKER_KUBECTL_PVC_ACCESS_MODE` The access mode for the pvc. Defaults to `ReadWriteOnce`.
1. `WOODPECKER_KUBECTL_PVC_STORAGE_CLASS` The storage class. By default empty and uses default.
1. `WOODPECKER_KUBECTL_PVC_ALLOW_ON_DETACHED` If true, allows detached jobs/pods to be mounted with a PVC. If used, you **must** provide a pvc storage class that allows for [ReadWriteMany](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) (for example NFS).

#### Advanced

1. `WOODPECKER_KUBECTL_REQUEST_TIMEOUT` The kubectl request timeout. Applies in,
   1. Waits for commands to complete.
   1. The kuectl `--request-timeout` if `WOODPECKER_KUBECTL_ALLOW_KUBECTL_CLIENT_CONFIG=true`
1. `WOODPECKER_KUBECTL_CONTAINER_START_DELAY` The delay time, in seconds, before the container starts.
1. `WOODPECKER_KUBECTL_ALLOW_KUBECTL_CLIENT_CONFIG` If true, allows the backend to configure the kubernetes client options (Relates to this [error](https://github.com/kubernetes/kubernetes/issues/93474)).
1. `WOODPECKER_KUBECTL_COMMAND_RETRIES_WAIT` The wait time between kubectl commands that may fail.
1. `WOODPECKER_KUBECTL_COMMAND_RETRIES` The number of retries for kubectl commands that may fail.
