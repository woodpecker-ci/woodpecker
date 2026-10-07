---
availability:
- {place: 'web::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/api/pipeline.go#GetPipelineMetadata}
- {kind: code, role: implementation, target: web/src/views/repo/pipeline/PipelineDebug.vue}
---

# Download pipeline metadata

A person with push access downloads a pipeline's metadata from its debug information, to run the pipeline again on their own machine.
