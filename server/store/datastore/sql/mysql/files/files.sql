-- name: files-find-build

SELECT
 file_id
,file_build_id
,file_proc_id
,file_pid
,file_name
,file_mime
,file_size
,file_time
,file_meta_passed
,file_meta_failed
,file_meta_skipped
FROM files
WHERE file_build_id = ?

-- name: files-find-proc-name

SELECT
 file_id
,file_build_id
,file_proc_id
,file_pid
,file_name
,file_mime
,file_size
,file_time
,file_meta_passed
,file_meta_failed
,file_meta_skipped
FROM files
WHERE file_proc_id = ?
  AND file_name    = ?

-- name: files-find-proc-name-data

SELECT
 file_id
,file_build_id
,file_proc_id
,file_pid
,file_name
,file_mime
,file_size
,file_time
,file_meta_passed
,file_meta_failed
,file_meta_skipped
,file_data
FROM files
WHERE file_proc_id = ?
  AND file_name    = ?

-- name: files-delete-build

DELETE FROM files WHERE file_build_id = ?
