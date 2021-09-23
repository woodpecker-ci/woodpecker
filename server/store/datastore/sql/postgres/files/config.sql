-- name: config-find-id

SELECT
 config.config_id
,config_repo_id
,config_hash
,config_data
,config_name
FROM config
LEFT JOIN build_config ON config.config_id = build_config.config_id
WHERE build_config.build_id = $1

-- name: config-find-repo-hash

SELECT
 config_id
,config_repo_id
,config_hash
,config_data
,config_name
FROM config
WHERE config_repo_id = $1
  AND config_hash    = $2

-- name: config-find-approved

SELECT build_id FROM builds
WHERE build_repo_id = $1
AND build_id in (
  SELECT build_id
  FROM build_config
  WHERE build_config.config_id = $2
  )
AND build_status NOT IN ('blocked', 'pending')
LIMIT 1
