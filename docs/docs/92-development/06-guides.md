# Guides

## Add new migration

Woodpecker uses migrations to change the database schema if a database model has been changed. If a developer for example adds a new property `IsKingKong` to the database model of a User in `server/model/` they would need to add a new migration like the following example to a file like `server/store/datastore/migration/123_add_is_king_kong_to_users.go`:

```go
// Copyright 2021 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"xorm.io/xorm"
)

var addIsKingKongToUsers = task{
	name: "add-king-kong-to-users",
	fn: func(sess *xorm.Session) error {
    // TODO
		return sess.Commit()
	},
}
```

:::tip
Woodpecker uses [Xorm](https://gitea.com/xorm/xorm) as ORM for the database connection. The `sess *xorm.Session` can be used to alter your database. You **don't** have to call `sess.Commit()` at the end of your migration as submitting the transaction / session will be done by the migration manager. After a successful execution of that transaction the server will automatically add the migration to a list, so it wont be executed again on the next start.
:::


To automatically execute the migration after the start of the server, the new migration needs to be added to the end of `migrationTasks` in `server/store/datastore/migration/migration.go`.
