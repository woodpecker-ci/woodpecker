---
toc_max_heading_level: 2
---

# Tenki

The Tenki backend runs each workflow inside a [Tenki](https://tenki.cloud) sandbox — an ephemeral, isolated microVM in the cloud. One sandbox is created per workflow and every step of that workflow is executed inside it, so steps share the same filesystem (the workspace), similar to how the Docker backend shares a volume between step containers.

Because steps run in a cloud microVM instead of on the agent host, this backend provides strong isolation for untrusted pipelines while keeping the agent itself lightweight.

In order to use this backend, you need to download (or build) the [agent](https://github.com/woodpecker-ci/woodpecker/releases/latest), configure it and run it. Set `WOODPECKER_BACKEND=tenki` and provide an API key (see the environment variables below).

:::note
Steps run on Tenki's standard Linux base image, so the step `image:` field is ignored — all commands run in that base image. Only steps that provide `commands` are supported: [services](../../../20-usage/60-services.md), plugins and the default (image-based) clone step are not yet supported, and a step without commands is rejected rather than silently skipped. Clone your repository with explicit `git` commands in the pipeline for now.
:::

:::warning
This backend runs untrusted pipeline code inside a Tenki cloud microVM. Be aware that:

- The sandbox base image grants **passwordless `sudo`** (root inside the microVM); isolation relies on the microVM boundary, not on an unprivileged user.
- **Outbound network access is enabled by default** (`WOODPECKER_BACKEND_TENKI_ALLOW_OUTBOUND`) so repositories and dependencies can be fetched. Disable it or restrict egress at the Tenki network layer for untrusted pipelines.
- Step environment variables (including secrets) and workflow labels are sent to the Tenki API to run the sandbox — that is, to a third-party control plane.

:::

## Pipeline example

The workflow compiler inserts the default image-based clone step unless `skip_clone: true` is set, and this backend rejects image-based steps. So set `skip_clone: true` and clone the repository yourself with `git`. The `image` field is required by the schema but ignored by this backend, so any value works as a placeholder.

```yaml title=".woodpecker.yaml"
when:
  - event: [push, pull_request]

skip_clone: true

steps:
  clone:
    image: alpine # placeholder: ignored by the tenki backend
    commands:
      - git clone "$CI_REPO_CLONE_URL" .
      - git checkout "$CI_COMMIT_SHA"
  build:
    image: alpine # placeholder
    commands:
      - go build ./...
```

For **private repositories**, provide credentials to `git`. The generated step script writes a `~/.netrc` from the standard `CI_NETRC_MACHINE` / `CI_NETRC_USERNAME` / `CI_NETRC_PASSWORD` variables when they are set, so an HTTPS clone authenticates automatically; alternatively embed a token in the clone URL.

## API key

Create an API key (`tk_...`) in the Tenki dashboard under **API Keys** and pass it via `WOODPECKER_BACKEND_TENKI_API_KEY`. The project and workspace are resolved automatically from the key's identity; set them explicitly only if the key can access more than one.

## Step specific configuration

### Working directory

The backend runs commands as a non-root user inside the sandbox. The workflow's workspace directory is created automatically before each step, so pipelines work with the default workspace path without extra configuration.

## Environment variables

### BACKEND_TENKI_API_KEY

- Name: `WOODPECKER_BACKEND_TENKI_API_KEY`
- Default: none

API key used to authenticate with Tenki. Required. Also read from `TENKI_API_KEY` and `TENKI_AUTH_TOKEN`.

---

### BACKEND_TENKI_ENDPOINT

- Name: `WOODPECKER_BACKEND_TENKI_ENDPOINT`
- Default: Tenki production endpoint

Base URL of the Tenki API. Leave empty to use the default (production) endpoint.

---

### BACKEND_TENKI_PROJECT_ID

- Name: `WOODPECKER_BACKEND_TENKI_PROJECT_ID`
- Default: first project of the resolved workspace

Project to create sandboxes in. Auto-resolved from the API key identity when empty. Also read from `TENKI_PROJECT_ID`.

---

### BACKEND_TENKI_WORKSPACE_ID

- Name: `WOODPECKER_BACKEND_TENKI_WORKSPACE_ID`
- Default: first workspace of the identity

Workspace to scope sandboxes to. Auto-resolved from the API key identity when empty. Also read from `TENKI_WORKSPACE_ID`.

---

### BACKEND_TENKI_ALLOW_OUTBOUND

- Name: `WOODPECKER_BACKEND_TENKI_ALLOW_OUTBOUND`
- Default: `true`

Allow outbound network access from the sandbox. Needed to clone repositories and fetch dependencies.

---

### BACKEND_TENKI_CREATE_TIMEOUT

- Name: `WOODPECKER_BACKEND_TENKI_CREATE_TIMEOUT`
- Default: `3m`

Maximum time to wait for a sandbox to become ready.

---

### BACKEND_TENKI_MAX_DURATION

- Name: `WOODPECKER_BACKEND_TENKI_MAX_DURATION`
- Default: `1h`

Maximum lifetime of a workflow sandbox before it is reclaimed.

---

### BACKEND_TENKI_IDLE_TIMEOUT

- Name: `WOODPECKER_BACKEND_TENKI_IDLE_TIMEOUT`
- Default: match the max duration

Inactivity window after which a sandbox is auto-paused. Because running a step does not refresh the session's activity clock, this defaults to the max duration so a long-running step is never paused mid-run. Set a smaller value only if you want idle sandboxes reclaimed sooner.
