## Yes, it's a fork

This repository is a hard fork of the Drone CI system.

Forked at the `0.8.9` version https://github.com/drone/drone/commit/768ed784bd74b0e0c2d8d49c4c8b6dca99b25e96

## Why fork?

Drone has been an open-core project since many prior versions. With each source file indicating whether it is part of the Apache 2.0 licensed or the propritary enterprise license. In the 0.8 line the enterprise features were limited to features like autoscaling and secret vaults.

However in the 1.0 line, databases other than SQLite, TLS support and agent based horizontal scaling were also moved under the enterprise license. Limiting the open source version to single node, hobbyist deployments.

The above feature reductions and the lack of clear communication of what is part of the open-source version led to this fork.

## The focus of this fork

The focus of this fork is

- Github
- Kubernetes and VM based backends
- Linux/amd64
- Some really good features that Drone 1.0 introduced: multiple pipelines, cron triggers

## Why should you use this fork?

you shouldn't necessarily. Paying for Drone 1.0 is a fine choice.

Check the issues and releases of this project if you are evaluating this project.
Also you can check the devlog to get the nuances: https://laszlo.cloud/drone-oss-08-devlog-1
