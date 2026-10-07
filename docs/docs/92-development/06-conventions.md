# Conventions

## Database naming

Database tables are named plural, columns don't have any prefix.

Example: Model name `Agent` with table name `agents` and columns `id`, `name`.

## Go code

Go code is formatted by `make format` and linted by `make lint`. The linters are configured in `.golangci.yaml`. These rules are the ones new contributors run into most often:

- Import aliases are fixed. A package can only be imported with an alias if `.golangci.yaml` defines one for it, and then only with exactly this alias (e.g. `forge_types` for `server/forge/types`). If a new package needs an alias, add it to the list.
- `context.WithCancel` is forbidden, use `context.WithCancelCause` instead.
- `print`, `println`, `panic` and `log.Fatal()` are forbidden. Return an error or use the logger instead.
- Top-level comments start with a capital letter and end with a period.

## Breaking changes

Every change that users, admins or API clients have to react to needs an entry in the [migration guide](/migrations) (`docs/src/pages/migrations.md`). This includes deprecations. Add it to the `next` version, in the section of the affected group (user-facing, admin-facing or API changes).

Write one entry per change, do not combine multiple changes into one entry.

Changes of the pipeline configuration have to follow the [deprecation policy](./40-deprecations.md) too.
