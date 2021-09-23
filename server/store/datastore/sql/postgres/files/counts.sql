-- name: count-users

SELECT reltuples
FROM pg_class WHERE relname = 'users'

-- name: count-repos

SELECT count(1)
FROM repos
WHERE repo_active = true

-- name: count-builds

SELECT count(1)
FROM builds
