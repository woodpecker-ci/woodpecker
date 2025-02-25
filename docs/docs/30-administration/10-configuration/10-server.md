---
toc_max_heading_level: 3
---

# Server

## Forge and User configuration

Woodpecker does not have its own user registration. Users are provided by your [forge](../11-forges/11-overview.md) (using OAuth2). The registration is closed by default (`WOODPECKER_OPEN=false`). If the registration is open, any user with an account can log in to Woodpecker with the configured forge.

You can also restrict the registration:

- closed registration and manually managing users with the CLI `woodpecker-cli user`
- open registration and allowing certain admin users with the setting `WOODPECKER_ADMIN`

  ```ini
  WOODPECKER_OPEN=false
  WOODPECKER_ADMIN=john.smith,jane_doe
  ```

- open registration and filtering by organizational affiliation with the setting `WOODPECKER_ORGS`

  ```ini
  WOODPECKER_OPEN=true
  WOODPECKER_ORGS=dolores,dog-patch
  ```

Administrators should also be explicitly set in your configuration.

```ini
WOODPECKER_ADMIN=john.smith,jane_doe
```

## Repository configuration

Woodpecker works with the user's OAuth permissions on the forge. By default Woodpecker will synchronize all repositories the user has access to. Use the variable `WOODPECKER_REPO_OWNERS` to filter which repos should only be synchronized by GitHub users. Normally you should enter the GitHub name of your company here.

```ini
WOODPECKER_REPO_OWNERS=my_company,my_company_oss_github_user
```

## Databases

The default database engine of Woodpecker is an embedded SQLite database which requires zero installation or configuration. But you can replace it with a MySQL/MariaDB or PostgreSQL database. There are also some fundamentals to keep in mind:

- Woodpecker does not create your database automatically. If you are using the MySQL or Postgres driver you will need to manually create your database using `CREATE DATABASE`.

- Woodpecker does not perform data archival; it considered out-of-scope for the project. Woodpecker is rather conservative with the amount of data it stores, however, you should expect the database logs to grow the size of your database considerably.

- Woodpecker automatically handles database migration, including the initial creation of tables and indexes. New versions of Woodpecker will automatically upgrade the database unless otherwise specified in the release notes.

- Woodpecker does not perform database backups. This should be handled by separate third party tools provided by your database vendor of choice.

### SQLite

By default Woodpecker uses a SQLite database stored under `/var/lib/woodpecker/`. If using containers, you can mount a [data volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to persist the SQLite database.

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
+    volumes:
+      - woodpecker-server-data:/var/lib/woodpecker/
```

### MySQL/MariaDB

The below example demonstrates MySQL database configuration. See the official driver [documentation](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for configuration options and examples.
The minimum version of MySQL/MariaDB required is determined by the `go-sql-driver/mysql` - see [it's README](https://github.com/go-sql-driver/mysql#requirements) for more information.

```ini
WOODPECKER_DATABASE_DRIVER=mysql
WOODPECKER_DATABASE_DATASOURCE=root:password@tcp(1.2.3.4:3306)/woodpecker?parseTime=true
```

### PostgreSQL

The below example demonstrates Postgres database configuration. See the official driver [documentation](https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING) for configuration options and examples.
Please use Postgres versions equal or higher than **11**.

```ini
WOODPECKER_DATABASE_DRIVER=postgres
WOODPECKER_DATABASE_DATASOURCE=postgres://root:password@1.2.3.4:5432/postgres?sslmode=disable
```

## UI customization

Woodpecker supports custom JS and CSS files. These files must be present in the server's filesystem.
They can be backed in a Docker image or mounted from a ConfigMap inside a Kubernetes environment.
The configuration variables are independent of each other, which means it can be just one file present, or both.

```ini
WOODPECKER_CUSTOM_CSS_FILE=/usr/local/www/woodpecker.css
WOODPECKER_CUSTOM_JS_FILE=/usr/local/www/woodpecker.js
```

The examples below show how to place a banner message in the top navigation bar of Woodpecker.

```css title="woodpecker.css"
.banner-message {
  position: absolute;
  width: 280px;
  height: 40px;
  margin-left: 240px;
  margin-top: 5px;
  padding-top: 5px;
  font-weight: bold;
  background: red no-repeat;
  text-align: center;
}
```

```javascript title="woodpecker.js"
// place/copy a minified version of your preferred lightweight JavaScript library here ...
!(function () {
  'use strict';
  function e() {} /*...*/
})();

$().ready(function () {
  $('.app nav img').first().htmlAfter("<div class='banner-message'>This is a demo banner message :)</div>");
});
```
