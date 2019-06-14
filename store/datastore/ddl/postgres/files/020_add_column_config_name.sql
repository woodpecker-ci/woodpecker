-- name: alter-table-add-config-name

ALTER TABLE config ADD COLUMN config_name TEXT

-- name: update-table-set-config-name

UPDATE config SET config_name = 'drone'