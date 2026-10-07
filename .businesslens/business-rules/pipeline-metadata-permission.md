---
appliesTo:
- type: entity
  id: pipeline
  effect: reads
  facts:
  - Metadata
permits:
- related:
  - {verb: has, entity: repository}
  - {verb: grants, entity: repository-permission}
  - {verb: holds, entity: user}
  when:
  - {entity: repository-permission, fact: Push, is: true}
- actors:
  - administrator
references:
- kind: code
  role: implementation
  target: server/api/pipeline.go#GetPipelineMetadata
---

# Only people with push access read a pipeline's metadata

A pipeline's debug information and metadata are shown only to people with push access to its repository, and to administrators.
