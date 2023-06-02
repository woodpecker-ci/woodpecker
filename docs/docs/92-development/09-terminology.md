# Terminology

## Glossary

- **Woodpecker CI**: The project name around Woodpecker.
- **Woodpecker**: An open-source tool that executes [pipelines][Pipeline] on your code.
- **Pipeline**: A sequence of [workflows][Workflow] that are executed on the code. [Pipelines][Pipeline] are triggered by events.
- **Workflow**: A sequence of steps and services that are executed as part of a [pipeline][Pipeline]. Workflows are represented by YAML files. Each [workflow][Workflow] has its own isolated [workspace][Workspace], and often additional resources like a shared network (docker).
- **Steps**: Individual commands, actions or tasks within a [workflow][Workflow].
- **Code**: Refers to the files tracked by the version control system used by the [forge][Forge].
- **Repos**: Short for repositories, these are storage locations where code is stored.
- **Forge**: The hosting platform or service where the repositories are hosted.
- **Workspace**: A folder shared between all steps of a [workflow][Workflow] containing the repository and all the generated data from previous steps.
- **Event**: Triggers the execution of a [pipeline][Pipeline], such as a [forge][Forge] event like `push`, or `manual` triggered manually from the UI.
- **Commit**: A defined state of the code, usually associated with a version control system like Git.
- **Matrix**: A configuration option that allows the execution of [workflows][Workflow] for each value in the [matrix][Matrix].
- **Service**: A service is a step that is executed from the start of a [workflow][Workflow] until its end. It can be accessed by name via the network from other steps within the same [workflow][Workflow].
- **Plugins**: [Plugins][Plugin] are extensions that provide pre-defined actions or commands for a step in a [workflow][Workflow]. They can be configured via settings.
- **Container**: A lightweight and isolated environment where commands are executed in.
- **YAML File**: A file format used to define and configure [workflows][Workflow].
- **Dependency**: [Workflows][Workflow] can depend on each other, and if possible, they are executed in parallel.
- **Status**: Status refers to the outcome of a step or [workflow][Workflow] after it has been executed, determined by the internal command exit code. At the end of a [workflow][Workflow] its status it is send to the [forge][Forge].

## Terms

Sometimes there exist multiple terms that can be used for a thing, we try to define it here once and stick to it.

- environment variables `*_LINK` should be `*_URL`, also in code, use `URL()` instead of `Link` ([Vote](https://framadate.org/jVSQHwIGfJYy82IL))
- **Pipelines** were previously called **builds**
- **Steps** were previously called **jobs**

[Pipeline]:  ../20-usage/20-pipeline-syntax.md
[Workflow]:  ../20-usage/25-workflows.md
[Forge]:     ../30-administration/11-forges/10-overview.md
[Plugin]:    ../20-usage/51-plugins/10-plugins.md
[Workspace]: ../20-usage/20-pipeline-syntax.md#workspace
[Matrix]:    ../20-usage/30-matrix-pipelines.md
