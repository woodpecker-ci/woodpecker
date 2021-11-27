# Status Badges

Woodpecker has integrated support for repository status badges. These badges can be added to your website or project readme file to display the status of your code.

## Badge endpoint

```text
<scheme>://<hostname>/api/badges/<owner>/<repo>/status.svg
```

The status badge displays the status for the latest build to your default branch (e.g. master). You can customize the branch by adding the `branch` query parameter.

```diff
-<scheme>://<hostname>/api/badges/<owner>/<repo>/status.svg
+<scheme>://<hostname>/api/badges/<owner>/<repo>/status.svg?branch=<branch>
```

Please note status badges do not include pull request results, since the status of a pull request does not provide an accurate representation of your repository state.
