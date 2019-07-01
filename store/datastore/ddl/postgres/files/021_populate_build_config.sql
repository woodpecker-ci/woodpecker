-- name: populate-build-config

INSERT INTO build_config (config_id, build_id)
SELECT build_config_id, build_id FROM builds
