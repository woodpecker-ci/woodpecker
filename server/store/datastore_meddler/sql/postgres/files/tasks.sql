-- name: task-list

SELECT
 task_id
,task_data
,task_labels
,task_dependencies
,task_run_on
FROM tasks

-- name: task-delete

DELETE FROM tasks WHERE task_id = $1
