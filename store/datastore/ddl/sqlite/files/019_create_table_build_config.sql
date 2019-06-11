-- name: create-table-build-config

CREATE TABLE IF NOT EXISTS build_config (
 config_id       INTEGER NOT NULL
,build_id        INTEGER NOT NULL
,PRIMARY KEY (config_id, build_id)
,FOREIGN KEY (config_id) REFERENCES config (config_id)
,FOREIGN KEY (build_id) REFERENCES builds (build_id)
);
