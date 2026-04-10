# Guides

## ORM

Woodpecker uses [Xorm](https://xorm.io/) as ORM for the database connection.

## Add a new migration

Woodpecker uses migrations to change the database schema if a database model has been changed. Add the new migration task into `server/store/datastore/migration/`.

:::info
Adding new properties to models will be handled automatically by the underlying [ORM](#orm) based on the [struct field tags](https://stackoverflow.com/questions/10858787/what-are-the-uses-for-tags-in-go) of the model. If you add a completely new model, you have to add it to the `allBeans` variable at `server/store/datastore/migration/migration.go` to get a new table created.
:::

:::warning
You should not use `sess.Begin()`, `sess.Commit()` or `sess.Close()` inside a migration. Session / transaction handling will be done by the underlying migration manager.
:::

To automatically execute the migration after the start of the server, the new migration needs to be added to the end of `migrationTasks` in `server/store/datastore/migration/migration.go`. After a successful execution of that transaction the server will automatically add the migration to a list, so it won't be executed again on the next start.

## Constants of official images

All official default images, are saved in [shared/constant/constant.go](https://github.com/woodpecker-ci/woodpecker/blob/main/shared/constant/constant.go) and must be pinned by an exact tag.

## Building images locally

### Server

if you want to build locally without buildx, you can use the following command:

```sh
export TAG='username/repo:tag' # Your image name
docker build -t $TAG -f docker/Dockerfile.server .
```

If you want to build multi-arch images, you can use the following command with buildx:

```sh
export PLATFORMS='linux/amd64,linux/arm64' # supported 'linux/amd64,linux/arm/v7,linux/arm64,linux/ppc64le,linux/riscv64'
export TAG='username/repo:tag' # Your image name
docker buildx build . --platform $PLATFORMS -t $TAG -f docker/Dockerfile.server --push # This will push the image to the registry, use --load to load it only locally (only single arch allowed)
```

:::info
The server Dockerfile defaults to a `scratch` base image. To build an Alpine-based image instead, add `--build-arg TARGET_BASE=alpine` to your build command.
:::

### Agent

```sh
export PLATFORMS='linux/amd64,linux/arm64' # supported 'linux/386,linux/amd64,freebsd/amd64,openbsd/amd64,linux/arm/v6,linux/arm/v7,linux/arm64,openbsd/arm64,freebsd/arm64,linux/ppc64le,linux/riscv64,linux/s390x'
export TAG='username/repo:tag' # Your image name
docker buildx build . --platform $PLATFORMS -t $TAG -f docker/Dockerfile.agent --push # This will push the image to the registry, use --load to load it only locally (only single arch allowed)
```

### CLI

#### build the image

```sh
export PLATFORMS='linux/amd64,linux/arm64' # supported 'linux/386,linux/amd64,freebsd/amd64,openbsd/amd64,linux/arm/v6,linux/arm/v7,linux/arm64,openbsd/arm64,freebsd/arm64,linux/ppc64le,linux/riscv64,linux/s390x'
export TAG='username/repo:tag' # Your image name
docker buildx build . --platform $PLATFORMS -t $TAG -f docker/Dockerfile.cli --push # This will push the image to the registry, use --load to load it only locally (only single arch allowed)
```

#### build the CLI

```sh
make build-cli
```
