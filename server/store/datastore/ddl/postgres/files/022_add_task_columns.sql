-- name: alter-table-add-task-dependencies
ALTER TABLE tasks ADD COLUMN task_dependencies BYTEA

-- name: alter-table-add-task-run-on

ALTER TABLE tasks ADD COLUMN task_run_on BYTEA
