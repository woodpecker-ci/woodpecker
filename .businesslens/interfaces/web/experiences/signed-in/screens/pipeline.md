---
entities:
- {entity: repository, shows: [Full name, Allow deployments]}
- {entity: pipeline, shows: [Number, Event, Branch, Commit, Message, Author, Created, Duration, Changed files, Errors, Configuration, Reviewer, Cancel info, Restarted from, Woodpecker version, Metadata], collects: [Deployment target, Deployment task, Additional pipeline variables]}
- {entity: workflow, shows: [Name, Platform, Agent, Error]}
- {entity: step, shows: [Name, Exit code, Duration, Log]}
entryPoints:
- web: /repos/{repoId}/pipeline/{pipelineId}
references:
- kind: code
  role: implementation
  target: web/src/views/repo/pipeline/PipelineWrapper.vue
- kind: code
  role: implementation
  target: web/src/views/repo/pipeline/PipelineDebug.vue
---

# Pipeline

One pipeline: its workflows and steps with their logs, the configuration it used, its changed files, its errors and, for people with push access, its debug information.
