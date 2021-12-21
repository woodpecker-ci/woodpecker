# Guides

## ORM

Woodpecker uses [Xorm](https://xorm.io/) as ORM for the database connection.
You can find its documentation at [gobook.io/read/gitea.com/xorm](https://gobook.io/read/gitea.com/xorm/manual-en-US/).

## Add a new migration

Woodpecker uses migrations to change the database schema if a database model has been changed. If for example a developer removes a property `Counter` from the model `Repo` in `server/model/` they would need to add a new migration task like the following  example to a file like `server/store/datastore/migration/004_repos_drop_repo_counter.go`:

```go
package migration

import (
	"xorm.io/xorm"
)

var alterTableReposDropCounter = task{
	name: "alter-table-drop-counter",
	fn: func(sess *xorm.Session) error {
		return dropTableColumns(sess, "repos", "repo_counter")
	},
}
```

:::info
Adding new properties to models will be handled automatically by the underlying [ORM](#orm) based on the [struct field tags](https://stackoverflow.com/questions/10858787/what-are-the-uses-for-tags-in-go) of the model. If you add a completely new model, you have to add it to the `syncAll()` function at `server/store/datastore/migration/migration.go` to get a new table created.
:::

:::warning
You should not use `sess.Begin()`, `sess.Commit()` or `sess.Close()` inside a migration. Session / transaction handling will be done by the underlying migration manager.
:::

To automatically execute the migration after the start of the server, the new migration needs to be added to the end of `migrationTasks` in `server/store/datastore/migration/migration.go`. After a successful execution of that transaction the server will automatically add the migration to a list, so it wont be executed again on the next start.

