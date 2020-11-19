-- name: globalsecret-find

SELECT
 secret_id
,secret_name
,secret_value
,secret_images
,secret_events
,secret_conceal
,secret_skip_verify
FROM global_secrets

-- name: globalsecret-find-name

SELECT
secret_id
,secret_name
,secret_value
,secret_images
,secret_events
,secret_conceal
,secret_skip_verify
FROM global_secrets
WHERE secret_name = ?

-- name: globalsecret-delete

DELETE FROM global_secrets WHERE secret_id = ?
