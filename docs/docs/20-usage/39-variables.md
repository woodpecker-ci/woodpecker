# Variables

Woodpecker provides the ability to store named parameters external to the YAML configuration file, in a central variable store. These variables can be passed to individual steps of the pipeline at runtime. <!-- TODO: not runtime but pipeline parse time to be exact -->

Woodpecker provides three different levels to add variables to your pipeline. The following list shows the priority of the different levels. If a variable is defined in multiple levels, will be used following this priorities: Repository variables > Organization variables > Global variables.

1. **Repository variables**: They are available to all pipelines of an repository.
2. **Organization variables**: They are available to all pipelines of an organization.
3. **Global variables**: Can be configured by an instance admin.

## Usage

### Use variables in settings and environment

You can set an setting or environment value from variables using the `from_variable` syntax.

In this example, the variable named `repo_name` would be passed to the setting named `repo`, which will be available in the plugin as environment variable named `PLUGIN_REPO_NAME` (See [plugins](./51-plugins/20-creating-plugins.md#settings) for details), and to the environment variable `REPO_NAME_ENV`.

```diff
 steps:
   - name: docker
     image: my-plugin
+    environment:
+      REPO_NAME_ENV:
+        from_variable: repo_name
+    settings:
+      repo:
+        from_variable: repo_name
```

## Adding Variables

Variables are added to the Woodpecker in the UI or with the CLI.

### CLI Examples

Create the variable using default settings. The variable will be available to all steps in your pipeline workflows, on all events.

```bash
woodpecker-cli variable add \
  -repository octocat/hello-world \
  -name repo_name \
  -value <value>
```

Loading variables from file using curl `@` syntax. This is the recommended approach for loading variables from file to preserve newlines:

```diff
 woodpecker-cli variable add \
   -repository octocat/hello-world \
   -name example_uname \
+  -value @/proc/version
```
