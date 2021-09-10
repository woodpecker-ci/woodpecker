# docker.io/go-docker
Official Go SDK for Docker

# Dependency management tool is required

This repository describes its dependencies in a `Gopkg.toml` file as created by the [`dep`](https://github.com/golang/dep#setup) tool.

It also uses semantic versioning, and requires its users to use `dep`-compatible dependency management tools to ensure stability and avoid breaking changes.

The canonical import path is `docker.io/go-docker`.

Note: you may download it with `go get -d docker.io/go-docker`, but if you omit `-d`, you may have compile errors. Hence the `dep` approach is preferred.

## How to use `dep` in your project

You can use any tool that is compatible, but in the examples below we are using `dep`.

### Adding dependency to `vendor/`

```bash
$ cd $GOPATH/src/myproject
$ dep init 					# only if first time use
$ dep ensure -add docker.io/go-docker@v1    	# to use the latest version of v1.x.y
```

### Updating dependency

```bash
$ cd $GOPATH/src/myproject
$ edit Gopkg.toml
$ dep ensure
```

# Reference Documentation

[godoc.org/docker.io/go-docker](https://godoc.org/docker.io/go-docker)

# Issues

Feel free to open issues on the Github issue tracker.

