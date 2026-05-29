# Status Badges

Woodpecker has integrated support for repository status badges. These badges can be added to your website or project readme file to display the status of your code.

## Badge endpoint

```uri
<scheme>://<hostname>/api/badges/<repo-id>/status.svg
```

The status badge displays the status for the latest build to your default branch (e.g. main). You can customize the branch by adding the `branch` query parameter.

```diff
-<scheme>://<hostname>/api/badges/<repo-id>/status.svg
+<scheme>://<hostname>/api/badges/<repo-id>/status.svg?branch=<branch>
```

By default status badges do not include pull request results, since the status of a pull request does not provide an accurate representation of your repository state.
If you'd like to respect other or further events, you can add the `events` query parameter, otherwise the badge represents only the state of the last push event:

```diff
-<scheme>://<hostname>/api/badges/<repo-id>/status.svg
+<scheme>://<hostname>/api/badges/<repo-id>/status.svg?events=manual,cron
```
