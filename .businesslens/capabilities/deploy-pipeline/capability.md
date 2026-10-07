---
availability:
- {place: 'web::signed-in'}
- {place: 'cli::signed-in'}
- {place: 'api::signed-in'}
domain: pipelines
references:
- {kind: code, role: implementation, target: server/api/pipeline.go#PostPipeline}
- {kind: code, role: implementation, target: web/src/components/layout/popups/DeployPipelinePopup.vue}
- {kind: code, role: implementation, target: cli/pipeline/deploy/deploy.go}
---

# Deploy pipeline

A person with push access deploys a pipeline to a target environment: a new pipeline with the deployment event from the earlier one's commit and configuration, a deployment task and optional variables. The web UI offers Deploy on successful pipelines while the repository allows deployments; the server refuses any deployment while the repository does not.
