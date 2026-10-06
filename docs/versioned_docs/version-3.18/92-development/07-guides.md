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

```sh
### build web component
make vendor
cd web/
pnpm install --frozen-lockfile
pnpm build
cd ..

### define the platforms to build for (e.g. linux/amd64)
# (the | is not a typo here)
export PLATFORMS='linux|amd64'
make cross-compile-server

### build the image
docker buildx build --platform linux/amd64 -t username/repo:tag -f docker/Dockerfile.server.multiarch.rootless --push .
```

:::info
The `cross-compile-server` rule makes use of `xgo`, a go cross-compiler. You need to be on a `amd64` host to do this, as `xgo` is only available for `amd64` (see [xgo#213](https://github.com/techknowlogick/xgo/issues/213)).
You can try to use the `build-server` rule instead, however this one fails for some OS (e.g. macOS).
:::

### Agent

```sh
### build the agent
make build-agent

### build the image
docker buildx build --platform linux/amd64 -t username/repo:tag -f docker/Dockerfile.agent.multiarch --push .
```

### CLI

```sh
### build the CLI
make build-cli

### build the image
docker buildx build --platform linux/amd64 -t username/repo:tag -f docker/Dockerfile.cli.multiarch.rootless --push .
```
