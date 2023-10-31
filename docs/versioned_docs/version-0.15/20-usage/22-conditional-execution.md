# Conditional Step Execution

Woodpecker supports defining conditions for pipeline step by a `when` block. If all conditions in the `when` block evaluate to true the step is executed, otherwise it is skipped.

## `repo`

Example conditional execution by repository:

```diff
 pipeline:
   slack:
     image: plugins/slack
     settings:
       channel: dev
+    when:
+      repo: test/test
```

## `branch`

Example conditional execution by branch:

```diff
pipeline:
  slack:
    image: plugins/slack
    settings:
      channel: dev
+   when:
+     branch: master
```

> The step now triggers on master, but also if the target branch of a pull request is `master`. Add an event condition to limit it further to pushes on master only.

Execute a step if the branch is `master` or `develop`:

```diff
when:
  branch: [master, develop]
```

Execute a step if the branch starts with `prefix/*`:

```diff
when:
  branch: prefix/*
```

Execute a step using custom include and exclude logic:

```diff
when:
  branch:
    include: [ master, release/* ]
    exclude: [ release/1.0.0, release/1.1.* ]
```

## `event`

Execute a step if the build event is a `tag`:

```diff
when:
  event: tag
```

Execute a step if the build event is a `tag` created from the specified branch:

```diff
when:
  event: tag
+ branch: master
```

Execute a step for all non-pull request events:

```diff
when:
  event: [push, tag, deployment]
```

Execute a step for all build events:

```diff
when:
  event: [push, pull_request, tag, deployment]
```

## `status`

There are use cases for executing pipeline steps on failure, such as sending notifications for failed pipelines. Use the status constraint to execute steps even when the pipeline fails:

```diff
pipeline:
  slack:
    image: plugins/slack
    settings:
      channel: dev
+   when:
+     status: [ success, failure ]
```

## `platform`

Execute a step for a specific platform:

```diff
when:
  platform: linux/amd64
```

Execute a step for a specific platform using wildcards:

```diff
when:
  platform:  [ linux/*, windows/amd64 ]
```

## `environment`

Execute a step for deployment events matching the target deployment environment:

```diff
when:
  environment: production
  event: deployment
```

## `matrix`

Execute a step for a single matrix permutation:

```diff
when:
  matrix:
    GO_VERSION: 1.5
    REDIS_VERSION: 2.8
```

## `instance`

Execute a step only on a certain Woodpecker instance matching the specified hostname:

```diff
when:
  instance: stage.woodpecker.company.com
```

## `path`

:::info
This feature is currently only available for GitHub, GitLab and Gitea.
Pull requests aren't supported by gitea at the moment ([go-gitea/gitea#18228](https://github.com/go-gitea/gitea/pull/18228)).
Path conditions are ignored for tag events.
:::

Execute a step only on a pipeline with certain files being changed:

```diff
when:
  path: "src/*"
```

You can use [glob patterns](https://github.com/bmatcuk/doublestar#patterns) to match the changed files and specify if the step should run if a file matching that pattern has been changed `include` or if some files have **not** been changed `exclude`.

```diff
when:
  path:
    include: [ '.woodpecker/*.yml', '*.ini' ]
    exclude: [ '*.md', 'docs/**' ]
    ignore_message: "[ALL]"
```

**Hint:** Passing a defined ignore-message like `[ALL]` inside the commit message will ignore all path conditions.
