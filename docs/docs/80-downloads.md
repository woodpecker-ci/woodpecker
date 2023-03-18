# Downloads

## Which version of Woodpecker should I use?

Woodpecker is having two different kinds of releases: **stable** and **next**.

To find out more about the differences between the two releases, please read the [FAQ](/faq).

## Binaries & DEB, RPM

[Latest release](https://github.com/woodpecker-ci/woodpecker/releases/latest)

## Docker images

Image variants:
* The `latest` image is the latest stable release
* The `vX.X.X` images are stable releases
* The `vX.X` images are based on the latest patch version of a specific minor release (see [Semver](https://semver.org/))
* The `next` images are based on the current master branch and should not be used for production environments

``` bash
# server
docker pull woodpeckerci/woodpecker-server:latest
docker pull woodpeckerci/woodpecker-server:latest-alpine

# agent
docker pull woodpeckerci/woodpecker-agent:latest
docker pull woodpeckerci/woodpecker-agent:latest-alpine

# cli
docker pull woodpeckerci/woodpecker-cli:latest
docker pull woodpeckerci/woodpecker-cli:latest-alpine
```
