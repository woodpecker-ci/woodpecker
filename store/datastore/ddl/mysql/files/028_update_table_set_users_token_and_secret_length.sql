-- name: update-table-set-users-token-and-secret-length

ALTER TABLE users MODIFY user_token varchar(1000);
ALTER TABLE users MODIFY user_secret varchar(1000);
