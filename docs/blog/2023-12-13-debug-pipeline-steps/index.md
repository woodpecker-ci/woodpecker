---
title: '[Community] Debug pipeline steps'
description: Debug pipeline steps using sshx
slug: debug-pipeline-steps
authors:
  - name: anbraten
    url: https://github.com/anbraten
    image_url: https://github.com/anbraten.png
hide_table_of_contents: true
tags: [community, debug]
---

<!-- cspell:ignore sshx -->

Sometimes you want to debug a pipeline.
Therefore I recently discovered: <https://github.com/ekzhang/sshx>

A simple step like should allow you to debug:

```yaml
steps:
  - name: debug
    image: alpine
    commands:
      - curl -sSf https://sshx.io/get | sh && sshx
      #      ^
      #      └ This will open a remote terminal session and print the URL. It
      #        should take under a second.
```
