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
