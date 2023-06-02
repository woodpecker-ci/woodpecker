# Terminology

## Glossary

- **Woodpecker**: An open-source tool that executes pipelines on your code.
- **Pipeline**: A sequence of workflows that are executed on the code. Pipelines are triggered by events.
- **Workflow**: A sequence of steps and services that are executed as part of a pipeline. Workflows are represented by YAML files. Each workflow has its own isolated workspace, and often additional resources like a shared network (docker).
- **Steps**: Individual commands, actions or tasks within a workflow.
- **Code**: Refers to the files tracked by the version control system used by the forge.
- **Repos**: Short for repositories, these are storage locations where code is stored.
- **Forge**: The hosting platform or service where the repositories are hosted.
- **Workspace**: A folder shared between all steps of a workflow containing the repository and all the generated data from previous steps.
- **Event**: Triggers the execution of a pipeline, such as a forge event like `push`, or `manual` triggered manually from the UI.
- **Commit**: A defined state of the code, usually associated with a version control system like Git.
- **Matrix**: A configuration option that allows the execution of workflows for each value in the matrix.
- **Service**: A service is a step that is executed from the start of a workflow until its end. It can be accessed by name via the network from other steps within the same workflow.
- **Plugins**: Plugins are extensions that provide pre-defined actions or commands for a step in a workflow. They can be configured via settings.
- **Container**: A lightweight and isolated environment where commands are executed in.
- **YAML File**: A file format used to define and configure workflows.
- **Dependency**: Workflows can depend on each other, and if possible, they are executed in parallel.
- **Status**: Status refers to the outcome of a step or workflow after it has been executed, determined by the internal command exit code. At the end of a workflow its status it is send to the forge.

## Terms

Sometimes there exist multiple terms that can be used for a thing, we try to define it here once and stick to it.

- environment variables `*_LINK` should be `*_URL`, also in code, use `URL()` instead of `Link` ([Vote](https://framadate.org/jVSQHwIGfJYy82IL))
- **Pipelines** were previously called **builds**
- **Steps** were previously called **jobs**
