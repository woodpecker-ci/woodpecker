# Awesome Woodpecker

A curated list of awesome things related to Woodpecker CI.

If you have some missing resources, please feel free to [open a pull-request](https://github.com/woodpecker-ci/woodpecker/edit/main/docs/docs/92-awesome.md) and add them.

## Official Resources

- [Woodpecker CI pipeline configs](https://github.com/woodpecker-ci/woodpecker/tree/main/.woodpecker) - Complex setup containing different kind of pipelines
  - [Golang tests](https://github.com/woodpecker-ci/woodpecker/blob/main/.woodpecker/test.yaml)
  - [Typescript, eslint & Vue](https://github.com/woodpecker-ci/woodpecker/blob/main/.woodpecker/web.yaml)
  - [Docusaurus & publishing to GitHub Pages](https://github.com/woodpecker-ci/woodpecker/blob/main/.woodpecker/docs.yaml)
  - [Docker container building](https://github.com/woodpecker-ci/woodpecker/blob/main/.woodpecker/docker.yaml)

## Projects using Woodpecker

- [Woodpecker CI](https://github.com/woodpecker-ci/woodpecker/tree/main/.woodpecker) itself
- [All official plugins](https://github.com/woodpecker-ci?q=plugin&type=all)
- [dessalines/thumb-key](https://github.com/dessalines/thumb-key/blob/main/.woodpecker.yml) - Android Jetpack compose linting and building
- [Vieter](https://git.rustybever.be/vieter-v/vieter) - Archlinux/Pacman repository server & automated package build system
  - [Rieter](https://git.rustybever.be/Chewing_Bever/rieter) - Rewrite of the Vieter project in Rust
- [Alex](https://git.rustybever.be/Chewing_Bever/alex) - Minecraft server wrapper designed to automate backups & complement Docker installations

## Tools

- [Convert Drone CI pipelines to Woodpecker CI](https://codeberg.org/lafriks/woodpecker-pipeline-transform)
- [Ansible NAS](https://github.com/davestephens/ansible-nas/) - a homelab Ansible playbook that can set up Woodpecker CI and Gitea
- [picus](https://github.com/windsource/picus) - Picus connects to a Woodpecker CI server and creates an agent in the cloud when there are pending workflows.
- [Hetzner cloud](https://www.hetzner.com/cloud) based [Woodpecker compatible autoscaler](https://git.ljoonal.xyz/ljoonal/hetzner-ci-autoscaler) - Creates and destroys VPS instances based on the count of pending & running jobs.
- [woodpecker-lint](https://git.schmidl.dev/schtobia/woodpecker-lint) - A repository for linting a Woodpecker config file via pre-commit hook
- [Grafana Dashboard](https://github.com/Janik-Haag/woodpecker-grafana-dashboard) - A dashboard visualizing information exposed by the Woodpecker prometheus endpoint.
- [woodpecker-autoscaler](https://github.com/Lerentis/woodpecker-autoscaler) - Yet another Woodpecker autoscaler currently targeting [Hetzner cloud](https://www.hetzner.com/cloud) that works in parallel to other autoscaler implementations.

## Configuration Services

- [Dynamic Pipelines for Nix Flakes](https://github.com/pinpox/woodpecker-flake-pipeliner) - Define pipelines as Nix Flake outputs

## Pipelines

- [Collection of pipeline examples](https://codeberg.org/Codeberg-CI/examples)

## Posts & tutorials

- [Setup Gitea with Woodpecker CI](https://containers.fan/posts/setup-gitea-with-woodpecker-ci/)
- [Step-by-step guide to modern, secure and Open-source CI setup](https://devforth.io/blog/step-by-step-guide-to-modern-secure-ci-setup/)
- [Using Woodpecker CI for my static sites](https://jan.wildeboer.net/2022/07/Woodpecker-CI-Jekyll/)
- [Woodpecker CI @ Codeberg](https://www.sarkasti.eu/articles/post/woodpecker/)
- [Deploy Docker/Compose using Woodpecker CI](https://hinty.io/vverenko/deploy-docker-compose-using-woodpecker-ci/)
- [Installing Woodpecker CI in your personal homelab](https://pwa.io/articles/installing-woodpecker-in-your-homelab/)
- [Locally Cached Nix CI with Woodpecker](https://blog.kotatsu.dev/posts/2023-04-21-woodpecker-nix-caching/)
- [How to run Cypress auto-tests on Woodpecker CI and report results to Slack](https://devforth.io/blog/how-to-run-cypress-auto-tests-on-woodpecker-ci-and-report-results-to-slack/)
- [Quest For CICD - WoodpeckerCI](https://omaramin.me/posts/woodpecker/)
- [Getting started with Woodpecker CI](https://systeemkabouter.eu/getting-started-with-woodpecker-ci.html)
- [Installing gitea and woodpecker using binary packages](https://neelex.com/2023/03/26/Installing-gitea-using-binary-packages/)
- [Deploying mdbook to codeberg pages using woodpecker CI](https://www.markpitblado.me/blog/deploying-mdbook-to-codeberg-pages-using-woodpecker-ci/)
- [Deploy a Fly app with Woodpecker CI](https://joeroe.io/2024/01/09/deploy-fly-woodpecker-ci.html)
- [Ansible - using Woodpecker as an alternative to Semaphore](https://pat-s.me/ansible-using-woodpecker-as-an-alternative-to-semaphore/)
- [Simple selfhosted CI/CD with Woodpecker](https://xyquadrat.ch/blog/simple-ci-with-woodpecker/)
- [Notes to self on Woodpecker-CI](https://jpmens.net/2023/09/22/notes-to-self-on-woodpecker-ci/)
- [CI/CD with Woodpecker and Gitea](https://wilw.dev/blog/2023/04/23/woodpecker-ci/)

## Videos

- [Replace Ansible Semaphore with Woodpecker CI](https://www.youtube.com/watch?v=d610YPvCB0E)
- ["unexpected EOF" error when trying to pair Woodpecker CI served through the Caddy with Gitea](https://www.youtube.com/watch?v=n7Hyvt71Np0)
- [CICD Environment in Docker Swarm behind Caddy Server - Part 2 Woodpeckerci](https://www.youtube.com/watch?v=rkbw_k7JvS0)
- [How to Build & Publish Custom Docker Container using Gitea & Woodpecker behind Caddy Server | TUNEIT](https://www.youtube.com/watch?v=9m7DbgL1mNk)
- [Radicle Woodpecker CI Integration](https://www.youtube.com/watch?v=Ks1nbYLn4P8)
- [woodpecker-ci/woodpecker - Gource visualisation](https://www.youtube.com/watch?v=38JuakZ6m5s)

## Plugins

We have a separate [index](/plugins) for plugins.
