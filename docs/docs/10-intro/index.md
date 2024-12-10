# Welcome to Woodpecker

Woodpecker is a CI/CD tool. It is designed to be lightweight (< 200 MB memory consumption), simple to use and fast and can be used with many different Git providers and backends (docker, kubernetes, local).

## The "CI/CD" concept

CI/CD stands for "Continuous Integration” and “Continuous Deployment.”
It is a streamlined process that moves your code from development to production while performing various checks, tests, and routines along the way.
A standard CI/CD pipeline typically includes steps such as:

1. Running tests
2. Building the application
3. Deploying the application

RedHat has written an [article which explains the concept in more detail](https://www.redhat.com/en/topics/devops/what-is-ci-cd).

## Containers at the core

In contrast to other CI/CD applications, Woodpecker solely focuses on using containers for executing workflows.
If you are already using containers in your daily workflow, you'll for sure love Woodpecker.

## Convinced? Get started by deploying your own Woodpecker instance

Woodpecker is [pretty lightweight](../30-administration/00-getting-started.md#hardware-requirements) and can even run on a Raspberry Pi without much impact.
To set up your own Woodpecker instance, follow the [deployment guide](../30-administration/00-getting-started.md).
