---
entities:
- {entity: repository, shows: [Full name]}
- {entity: pipeline, shows: [Number, Event, Branch, Commit, Message, Author, Created, Duration, Changed files, Errors, Configuration, Cancel info, Woodpecker version]}
- {entity: workflow, shows: [Name, Platform, Agent, Error]}
- {entity: step, shows: [Name, Exit code, Duration, Log]}
entryPoints:
- web: /repos/{repoId}/pipeline/{pipelineId}
references:
- kind: code
  role: implementation
  target: web/src/views/repo/pipeline/PipelineWrapper.vue
---

# Pipeline

One pipeline of a public repository: its workflows and steps with their logs, the configuration it used, its changed files and its errors.
