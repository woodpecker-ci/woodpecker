-- name: update-table-set-users-token-and-secret-length

ALTER TABLE users ALTER COLUMN user_token TYPE varchar(1000);
ALTER TABLE users ALTER COLUMN user_secret TYPE varchar(1000);
