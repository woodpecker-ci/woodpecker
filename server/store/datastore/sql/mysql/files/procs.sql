-- name: procs-find-id

SELECT
 proc_id
,proc_build_id
,proc_pid
,proc_ppid
,proc_pgid
,proc_name
,proc_state
,proc_error
,proc_exit_code
,proc_started
,proc_stopped
,proc_machine
,proc_platform
,proc_environ
FROM procs
WHERE proc_id = ?

-- name: procs-find-build

SELECT
 proc_id
,proc_build_id
,proc_pid
,proc_ppid
,proc_pgid
,proc_name
,proc_state
,proc_error
,proc_exit_code
,proc_started
,proc_stopped
,proc_machine
,proc_platform
,proc_environ
FROM procs
WHERE proc_build_id = ?
ORDER BY proc_id ASC

-- name: procs-find-build-pid

SELECT
proc_id
,proc_build_id
,proc_pid
,proc_ppid
,proc_pgid
,proc_name
,proc_state
,proc_error
,proc_exit_code
,proc_started
,proc_stopped
,proc_machine
,proc_platform
,proc_environ
FROM procs
WHERE proc_build_id = ?
  AND proc_pid = ?

-- name: procs-find-build-ppid

SELECT
proc_id
,proc_build_id
,proc_pid
,proc_ppid
,proc_pgid
,proc_name
,proc_state
,proc_error
,proc_exit_code
,proc_started
,proc_stopped
,proc_machine
,proc_platform
,proc_environ
FROM procs
WHERE proc_build_id = ?
  AND proc_ppid = ?
  AND proc_name = ?

-- name: procs-delete-build

DELETE FROM procs WHERE proc_build_id = ?
