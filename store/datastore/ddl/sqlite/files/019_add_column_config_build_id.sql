-- name: alter-table-add-config-build-id

ALTER TABLE config ADD COLUMN config_build_id INTEGER

-- name: update-table-set-config-config-id

UPDATE config SET config_build_id = (SELECT builds.build_id FROM builds WHERE builds.build_config_id = config.config_id)
