-- name: alter-table-add-task-dependencies
ALTER TABLE tasks ADD COLUMN task_dependencies MEDIUMBLOB

-- name: alter-table-add-task-run-on

ALTER TABLE tasks ADD COLUMN task_run_on MEDIUMBLOB