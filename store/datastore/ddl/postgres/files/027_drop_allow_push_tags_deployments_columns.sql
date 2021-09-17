-- name: drop-allow-push-tags-deploys-columns
ALTER TABLE repo DROP COLUMN repo_allow_push, DROP COLUMN repo_allow_deploys, DROP COLUMN repo_allow_tags
