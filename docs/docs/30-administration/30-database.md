# Databases

The default database engine of Woodpecker is an embedded SQLite database which requires zero installation or configuration. But you can replace it with a MySQL/MariaDB or Postgres database.

## Configure SQLite

By default Woodpecker uses a SQLite database stored under `/var/lib/woodpecker/`. You can mount a [data volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to persist the SQLite database.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
+   volumes:
+     - woodpecker-server-data:/var/lib/woodpecker/
```

## Configure MySQL/MariaDB

The below example demonstrates MySQL database configuration. See the official driver [documentation](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for configuration options and examples.
The minimum version of MySQL/MariaDB required is determined by the `go-sql-driver/mysql` - see [it's README](https://github.com/go-sql-driver/mysql#requirements) for more information.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
+     - WOODPECKER_DATABASE_DRIVER=mysql
+     - WOODPECKER_DATABASE_DATASOURCE=root:password@tcp(1.2.3.4:3306)/woodpecker?parseTime=true
```

## Configure Postgres

The below example demonstrates Postgres database configuration. See the official driver [documentation](https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING) for configuration options and examples.
Please use Postgres versions equal or higher than **11**.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
+     - WOODPECKER_DATABASE_DRIVER=postgres
+     - WOODPECKER_DATABASE_DATASOURCE=postgres://root:password@1.2.3.4:5432/postgres?sslmode=disable
```

## Database Creation

If you use SQLite the file is created automatically.
Woodpecker does **not* create your database automatically, if you are using the MySQL or Postgres. You will need to manually create your database using `CREATE DATABASE`.

## Database Schema Migration

Woodpecker automatically handles database migration between versions, including the initial creation of tables and indexes. New versions of Woodpecker will automatically upgrade the database unless otherwise specified in the release notes.

## Database DBMS Migration

If you have an existing database and want to change the DBMS (e.g. from SQLite to Mariadb):

1. Rename `WOODPECKER_DATABASE_DRIVER` to `WOODPECKER_OLD_DATABASE_DRIVER` and `WOODPECKER_DATABASE_DATASOURCE` to `WOODPECKER_OLD_DATABASE_DATASOURCE`
2. Add the database configuration as you would with a new installation.  

On next start, a schema migration will run on the **old** database.
Then the new database is initialized and all data copied.

:::info
If you don't want to start the server afterwards set `WOODPECKER_OLD_DATABASE_IMPORT_ONLY` to **true**.  
If the new database already contains data the server will just error and exit.
:::
## Database Backups

Woodpecker does not perform database backups. This should be handled by separate third party tools provided by your database vendor of choice.

## Database Archiving

Woodpecker does not perform data archival; it considered out-of-scope for the project. Woodpecker is rather conservative with the amount of data it stores, however, you should expect the database logs to grow the size of your database considerably.
