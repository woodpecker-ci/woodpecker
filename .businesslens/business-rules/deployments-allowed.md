---
appliesTo:
- type: capability
  id: deploy-pipeline
- type: entity
  id: repository
  facts:
  - Allow deployments
references:
- kind: code
  role: implementation
  target: server/api/pipeline.go#PostPipeline
---

# A pipeline is deployed only while its repository allows deployments

Deploy is offered only for successful pipelines of repositories that allow deployments, and refused otherwise.
