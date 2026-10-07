---
scope: The Woodpecker server, its REST API and web UI, and the CLI as they shape product behavior.
method: Static inspection of source, UI strings and documentation; nothing was run.
covered:
- description: Server HTTP routing, permission middleware and API handlers.
  paths:
  - server/router/
  - server/api/
- description: Pipeline creation, approval, cancellation, restart and status handling.
  paths:
  - server/pipeline/
- description: Cron scheduler, queue and scheduler on the server.
  paths:
  - server/cron/
  - server/queue/
  - server/scheduler/
- description: Server domain model.
  paths:
  - server/model/
- description: Server services for secrets, registries, configuration, extensions and permissions.
  paths:
  - server/services/config/
  - server/services/secret/
  - server/services/registry/
  - server/services/permissions/
- description: Status badge and CCMenu rendering.
  paths:
  - server/badges/
  - server/ccmenu/
- description: Web UI views, routing, compositions and translations.
  paths:
  - web/src/
- description: CLI commands.
  paths:
  - cli/
  - cmd/cli/
- description: Secret filtering and log masking in the compiler, the linter and the agent's logger.
  paths:
  - pipeline/frontend/yaml/compiler/
  - pipeline/frontend/yaml/linter/
  - pipeline/shared/
  - agent/logger.go
- description: Server flags that set product behavior and the API description.
  paths:
  - cmd/server/flags.go
  - cmd/server/openapi.go
exclusions:
- description: Documentation website and blog.
  paths:
  - docs/
- description: Build, packaging, release and CI configuration.
  paths:
  - docker/
  - nfpm/
  - .woodpecker/
  - .github/
  - Makefile
  - flake.nix
  - release-config.ts
  - tools/
  - contrib/
- description: End-to-end test harness.
  paths:
  - e2e/
- description: Database storage and migrations.
  paths:
  - server/store/
- description: Go client library the CLI calls the API with.
  paths:
  - woodpecker-go/
- description: Generated agent transport definitions.
  paths:
  - rpc/proto/
unmapped:
- description: 'Pipeline configuration syntax: conditions, matrix, services, plugins, workspace and clone.'
  paths:
  - pipeline/frontend/builder/
  - pipeline/frontend/metadata/
  - pipeline/frontend/yaml/constraint/
  - pipeline/frontend/yaml/types/
- description: Agent process, runtime and container, Kubernetes and local backends, and the agent connection protocol.
  paths:
  - agent/
  - cmd/agent/
  - pipeline/backend/
  - pipeline/runtime/
  - server/rpc/
- description: Server start-up, setup and remaining configuration flags.
  paths:
  - cmd/server/
- description: 'Server diagnostics: metrics, health and profiling.'
  paths:
  - server/metric/
  - server/api/metrics/
  - server/api/debug/
  - server/logging/
- description: Live event feed the web UI subscribes to.
  paths:
  - server/pubsub/
  - server/api/stream.go
- description: CLI contexts, self-update and user info.
  paths:
  - cli/context/
  - cli/update/
  - cli/info/
- description: Log storage and log addons.
  paths:
  - server/services/log/
- description: Organization lookup and permission endpoints, and repository branch and pull request lists.
  paths:
  - server/api/org.go
  - server/api/repo.go
limitations:
- description: 'Forge drivers: how each forge maps its events, statuses, teams and permissions.'
  paths:
  - server/forge/
---

# Coverage
