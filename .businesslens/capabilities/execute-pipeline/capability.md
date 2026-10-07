---
availability:
- {place: 'web::public'}
- {place: 'web::signed-in'}
domain: pipelines
references:
- kind: code
  role: implementation
  target: server/rpc/rpc.go#RPC.Next
- kind: code
  role: implementation
  target: server/rpc/rpc.go#RPC.Done
- kind: code
  role: implementation
  target: server/scheduler/filter.go
- kind: code
  role: implementation
  target: server/queue/fifo.go
- kind: code
  role: implementation
  target: pipeline/frontend/yaml/compiler/compiler.go#Secret.Available
- kind: doc
  role: context
  target: docs/docs/20-usage/15-terminology/index.md
---

# Execute pipeline

Woodpecker gives queued workflows to agents whose labels and organization match, runs their steps with the secrets and registries they may receive, streams the logs and reports each pipeline's result to the forge.
