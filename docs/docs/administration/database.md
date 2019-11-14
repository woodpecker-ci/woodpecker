# Databases

This guide provides instructions for using alternate storage engines. Please note this is optional. The default storage engine is an embedded SQLite database which requires zero installation or configuration.

## Configure MySQL

The below example demonstrates mysql database configuration. See the official driver [documentation](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for configuration options and examples.

```diff
version: '3'

services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
+     DRONE_DATABASE_DRIVER: mysql
+     DRONE_DATABASE_DATASOURCE: root:password@tcp(1.2.3.4:3306)/drone?parseTime=true
```

## Configure Postgres

The below example demonstrates postgres database configuration. See the official driver [documentation](https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING) for configuration options and examples.

```diff
version: '3'

services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
+     DRONE_DATABASE_DRIVER: postgres
+     DRONE_DATABASE_DATASOURCE: postgres://root:password@1.2.3.4:5432/postgres?sslmode=disable
```

## Database Creation

Woodpecker does not create your database automatically. If you are using the mysql or postgres driver you will need to manually create your database using `CREATE DATABASE`

## Database Migration

Woodpecker automatically handles database migration, including the initial creation of tables and indexes. New versions of Woodpecker will automatically upgrade the database unless otherwise specified in the release notes.

## Database Backups

Woodpecker does not perform database backups. This should be handled by separate third party tools provided by your database vendor of choice.

## Database Archiving

Woodpecker does not perform data archival; it considered out-of-scope for the project. Woodpecker is rather conservative with the amount of data it stores, however, you should expect the database logs to grow the size of your database considerably.
