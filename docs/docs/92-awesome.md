# Awesome Woodpecker

A curated list of awesome things related to Woodpecker-CI.

If you have some missing resources, please feel free to [open a pull-request](https://github.com/woodpecker-ci/woodpecker/edit/master/docs/docs/92-awesome.md) and add them.

## Official Resources

- [Woodpecker CI pipeline configs](https://github.com/woodpecker-ci/woodpecker/tree/master/.woodpecker) - Complex setup containing different kind of pipelines
  - [Golang tests](https://github.com/woodpecker-ci/woodpecker/blob/master/.woodpecker/test.yml)
  - [Typescript, eslint & Vue](https://github.com/woodpecker-ci/woodpecker/blob/master/.woodpecker/web.yml)
  - [Docusaurus & publishing to GitHub Pages](https://github.com/woodpecker-ci/woodpecker/blob/master/.woodpecker/docs.yml)
  - [Docker container building](https://github.com/woodpecker-ci/woodpecker/blob/master/.woodpecker/docker.yml)
  - [Helm chart linting & releasing](https://github.com/woodpecker-ci/woodpecker/blob/master/.woodpecker/helm.yml)
  - [Android Jetpack compose linting and building](https://github.com/dessalines/thumb-key/blob/main/.woodpecker.yml)

## Projects using Woodpecker

- [Woodpecker-CI](https://github.com/woodpecker-ci/woodpecker/tree/master/.woodpecker) itself
- [All official plugins](https://github.com/woodpecker-ci?q=plugin&type=all)

## Tools

- [Convert Drone CI pipelines to Woodpecker CI](https://codeberg.org/lafriks/woodpecker-pipeline-transform)
- [Ansible NAS](https://github.com/davestephens/ansible-nas/) - a homelab Ansible playbook that can set up Woodpecker-CI and Gitea
- [picus](https://github.com/windsource/picus) - Picus connects to a Woodpecker CI server and creates an agent in the cloud when there are pending workflows.
- [Hetzner cloud](https://www.hetzner.com/cloud) based [Woodpecker compatible autoscaler](https://git.ljoonal.xyz/ljoonal/hetzner-ci-autoscaler) - Creates and destroys VPS instances based on the count of pending & running jobs.

## Templates

## Pipelines

- [Collection of pipeline examples](https://codeberg.org/Codeberg-CI/examples)

## Posts & tutorials

- [Setup Gitea with Woodpecker CI](https://containers.fan/posts/setup-gitea-with-woodpecker-ci/)
- [Step-by-step guide to modern, secure and Open-source CI setup](https://devforth.io/blog/step-by-step-guide-to-modern-secure-ci-setup/)
- [Using Woodpecker CI for my static sites](https://jan.wildeboer.net/2022/07/Woodpecker-CI-Jekyll/)
- [Woodpecker CI @ Codeberg](https://www.sarkasti.eu/articles/post/woodpecker/)
- [Deploy Docker/Compose using Woodpecker CI](https://hinty.io/vverenko/deploy-docker-compose-using-woodpecker-ci/)

## Videos

## Plugins

We have a separate [index](/plugins) for plugins.
