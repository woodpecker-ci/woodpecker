-- name: config-find-id

SELECT
 config_id
,config_repo_id
,config_build_id
,config_hash
,config_data
,config_name
FROM config
WHERE config_build_id = ?

-- name: config-find-repo-hash

SELECT
 config_id
,config_repo_id
,config_build_id
,config_hash
,config_data
,config_name
FROM config
WHERE config_repo_id = ?
  AND config_hash    = ?

-- name: config-find-approved

SELECT build_id FROM builds
WHERE build_repo_id = ?
AND build_id in (SELECT config_build_id FROM config WHERE config.config_id = ?)
AND build_status NOT IN ('blocked', 'pending')
LIMIT 1
