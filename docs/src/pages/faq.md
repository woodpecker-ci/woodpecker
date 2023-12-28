# FAQ

## What are the differences to Drone?

Apart from Woodpecker staying free and OpenSource forever, the growing community already introduced some nifty features like:

- [Multiple workflows](/docs/next/usage/workflows)
- [Conditional step execution on file changes](/docs/next/usage/workflow-syntax#path)
- [More features are already in the pipeline :wink:](https://github.com/woodpecker-ci/woodpecker/pulls) ...

## Why is Woodpecker a fork of Drone version 0.8?

The Drone CI license was changed after the 0.8 release from Apache 2 to a proprietary license. Woodpecker is based on this latest freely available version.

## Which version of Woodpecker should I use?

Woodpecker is having two different kinds of releases: **stable** and **next**.

The **stable** releases (currently version 2.1) are long-term supported (LTS) stable versions. The stable releases are only getting bugfixes.

The **next** release contains all bugfixes and features from `main` branch. Normally it should be pretty stable, but as its frequently updated, it might contain some bugs from time to time. There are no binaries for this version.

If you want all (new) features of Woodpecker and are willing to accept some possible bugs from time to time, you should use the next release otherwise use the stable release.

## How to debug clone issues

(And what to do with an error message like `fatal: could not read Username for 'https://<url>': No such device or address`)

This error can have multiple causes. If you use internal repositories you might have to enable `WOODPECKER_AUTHENTICATE_PUBLIC_REPOS`:

```yaml
services:
  woodpecker-server:
    [...]
    environment:
      - [...]
      - WOODPECKER_AUTHENTICATE_PUBLIC_REPOS=true
```

If that does not work, try to make sure the container can reach your git server. In order to do that disable git checkout and make the container "hang":

```yaml
skip_clone: true

steps:
  build:
    image: debian:stable-backports
    commands:
      - apt update
      - apt install -y inetutils-ping wget
      - ping -c 4 git.example.com
      - wget git.example.com
      - sleep 9999999
```

Get the container id using `docker ps` and copy the id from the first column. Enter the container with: `docker exec -it 1234asdf  bash` (replace `1234asdf` with the docker id). Then try to clone the git repository with the commands from the failing pipeline:

```bash
git init
git remote add origin https://git.example.com/username/repo.git
git fetch --no-tags origin +refs/heads/branch:
```

(replace the url AND the branch with the correct values, use your username and password as log in values)
