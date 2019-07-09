-- name: alter-table-add-task-dependencies
ALTER TABLE tasks ADD COLUMN task_dependencies BLOB

-- name: alter-table-add-task-run-on

ALTER TABLE tasks ADD COLUMN task_run_on BLOB
