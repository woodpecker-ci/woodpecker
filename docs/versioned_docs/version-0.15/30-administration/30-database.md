# Databases

The default database engine of Woodpecker is an embedded SQLite database which requires zero installation or configuration. But you can replace it with a MySQL or Postgres database.

## Configure sqlite

By default Woodpecker uses a sqlite database stored under `/var/lib/woodpecker/`. You can mount a [data volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to persist the sqlite database.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
+   volumes:
+     - woodpecker-server-data:/var/lib/woodpecker/
```

## Configure MySQL

The below example demonstrates mysql database configuration. See the official driver [documentation](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for configuration options and examples.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
+     WOODPECKER_DATABASE_DRIVER: mysql
+     WOODPECKER_DATABASE_DATASOURCE: root:password@tcp(1.2.3.4:3306)/woodpecker?parseTime=true
```

## Configure Postgres

The below example demonstrates postgres database configuration. See the official driver [documentation](https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING) for configuration options and examples.
Please use postgres versions equal or higher than **11**.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
+     WOODPECKER_DATABASE_DRIVER: postgres
+     WOODPECKER_DATABASE_DATASOURCE: postgres://root:password@1.2.3.4:5432/postgres?sslmode=disable
```

## Database Creation

Woodpecker does not create your database automatically. If you are using the mysql or postgres driver you will need to manually create your database using `CREATE DATABASE`

## Database Migration

Woodpecker automatically handles database migration, including the initial creation of tables and indexes. New versions of Woodpecker will automatically upgrade the database unless otherwise specified in the release notes.

## Database Backups

Woodpecker does not perform database backups. This should be handled by separate third party tools provided by your database vendor of choice.

## Database Archiving

Woodpecker does not perform data archival; it considered out-of-scope for the project. Woodpecker is rather conservative with the amount of data it stores, however, you should expect the database logs to grow the size of your database considerably.
