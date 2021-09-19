package postgres

// Lookup returns the named statement.
func Lookup(name string) string {
	return index[name]
}

var index = map[string]string{
	"config-find-id":              configFindId,
	"config-find-repo-hash":       configFindRepoHash,
	"config-find-approved":        configFindApproved,
	"count-users":                 countUsers,
	"count-repos":                 countRepos,
	"count-builds":                countBuilds,
	"feed-latest-build":           feedLatestBuild,
	"feed":                        feed,
	"files-find-build":            filesFindBuild,
	"files-find-proc-name":        filesFindProcName,
	"files-find-proc-name-data":   filesFindProcNameData,
	"files-delete-build":          filesDeleteBuild,
	"logs-find-proc":              logsFindProc,
	"perms-find-user":             permsFindUser,
	"perms-find-user-repo":        permsFindUserRepo,
	"perms-insert-replace":        permsInsertReplace,
	"perms-insert-replace-lookup": permsInsertReplaceLookup,
	"perms-delete-user-repo":      permsDeleteUserRepo,
	"perms-delete-user-date":      permsDeleteUserDate,
	"procs-find-id":               procsFindId,
	"procs-find-build":            procsFindBuild,
	"procs-find-build-pid":        procsFindBuildPid,
	"procs-find-build-ppid":       procsFindBuildPpid,
	"procs-delete-build":          procsDeleteBuild,
	"registry-find-repo":          registryFindRepo,
	"registry-find-repo-addr":     registryFindRepoAddr,
	"registry-delete-repo":        registryDeleteRepo,
	"registry-delete":             registryDelete,
	"repo-update-counter":         repoUpdateCounter,
	"repo-find-user":              repoFindUser,
	"repo-insert-ignore":          repoInsertIgnore,
	"repo-delete":                 repoDelete,
	"secret-find-repo":            secretFindRepo,
	"secret-find-repo-name":       secretFindRepoName,
	"secret-delete":               secretDelete,
	"sender-find-repo":            senderFindRepo,
	"sender-find-repo-login":      senderFindRepoLogin,
	"sender-delete-repo":          senderDeleteRepo,
	"sender-delete":               senderDelete,
	"task-list":                   taskList,
	"task-delete":                 taskDelete,
	"user-find":                   userFind,
	"user-find-login":             userFindLogin,
	"user-update":                 userUpdate,
	"user-delete":                 userDelete,
}

var configFindId = `
SELECT
 config.config_id
,config_repo_id
,config_hash
,config_data
,config_name
FROM config
LEFT JOIN build_config ON config.config_id = build_config.config_id
WHERE build_config.build_id = $1
`

var configFindRepoHash = `
SELECT
 config_id
,config_repo_id
,config_hash
,config_data
,config_name
FROM config
WHERE config_repo_id = $1
  AND config_hash    = $2
`

var configFindApproved = `
SELECT build_id FROM builds
WHERE build_repo_id = $1
AND build_id in (
  SELECT build_id
  FROM build_config
  WHERE build_config.config_id = $2
  )
AND build_status NOT IN ('blocked', 'pending')
LIMIT 1
`

var countUsers = `
SELECT reltuples
FROM pg_class WHERE relname = 'users'
`

var countRepos = `
SELECT count(1)
FROM repos
WHERE repo_active = true
`

var countBuilds = `
SELECT count(1)
FROM builds
`

var feedLatestBuild = `
SELECT
 repo_owner
,repo_name
,repo_full_name
,build_number
,build_event
,build_status
,build_created
,build_started
,build_finished
,build_commit
,build_branch
,build_ref
,build_refspec
,build_remote
,build_title
,build_message
,build_author
,build_email
,build_avatar
FROM repos LEFT OUTER JOIN (
	SELECT DISTINCT ON (build_repo_id) * FROM builds
	ORDER BY build_repo_id, build_id DESC
) b ON b.build_repo_id = repos.repo_id
INNER JOIN perms ON perms.perm_repo_id = repos.repo_id
WHERE perms.perm_user_id = $1
  AND repos.repo_active = TRUE
ORDER BY repo_full_name ASC;
`

var feed = `
SELECT
 repo_owner
,repo_name
,repo_full_name
,build_number
,build_event
,build_status
,build_created
,build_started
,build_finished
,build_commit
,build_branch
,build_ref
,build_refspec
,build_remote
,build_title
,build_message
,build_author
,build_email
,build_avatar
FROM repos
INNER JOIN perms  ON perms.perm_repo_id   = repos.repo_id
INNER JOIN builds ON builds.build_repo_id = repos.repo_id
WHERE perms.perm_user_id = $1
ORDER BY build_id DESC
LIMIT 50
`

var filesFindBuild = `
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
WHERE file_build_id = $1
`

var filesFindProcName = `
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
WHERE file_proc_id = $1
  AND file_name    = $2
`

var filesFindProcNameData = `
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
WHERE file_proc_id = $1
  AND file_name    = $2
`

var filesDeleteBuild = `
DELETE FROM files WHERE file_build_id = $1
`

var logsFindProc = `
SELECT
 log_id
,log_job_id
,log_data
FROM logs
WHERE log_job_id = $1
LIMIT 1
`

var permsFindUser = `
SELECT
 perm_user_id
,perm_repo_id
,perm_pull
,perm_push
,perm_admin
,perm_date
FROM perms
WHERE perm_user_id = $1
`

var permsFindUserRepo = `
SELECT
 perm_user_id
,perm_repo_id
,perm_pull
,perm_push
,perm_admin
,perm_synced
FROM perms
WHERE perm_user_id = $1
  AND perm_repo_id = $2
`

var permsInsertReplace = `
REPLACE INTO perms (
 perm_user_id
,perm_repo_id
,perm_pull
,perm_push
,perm_admin
,perm_synced
) VALUES ($1,$2,$3,$4,$5,$6)
`

var permsInsertReplaceLookup = `
INSERT INTO perms (
 perm_user_id
,perm_repo_id
,perm_pull
,perm_push
,perm_admin
,perm_synced
) VALUES ($1,(SELECT repo_id FROM repos WHERE repo_full_name = $2),$3,$4,$5,$6)
ON CONFLICT (perm_user_id, perm_repo_id) DO UPDATE SET
 perm_pull = EXCLUDED.perm_pull
,perm_push = EXCLUDED.perm_push
,perm_admin = EXCLUDED.perm_admin
,perm_synced = EXCLUDED.perm_synced
`

var permsDeleteUserRepo = `
DELETE FROM perms
WHERE perm_user_id = $1
  AND perm_repo_id = $2
`

var permsDeleteUserDate = `
DELETE FROM perms
WHERE perm_user_id = $1
  AND perm_synced < $2
`

var procsFindId = `
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
WHERE proc_id = $1
`

var procsFindBuild = `
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
WHERE proc_build_id = $1
ORDER BY proc_id ASC
`

var procsFindBuildPid = `
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
WHERE proc_build_id = $1
  AND proc_pid      = $2
`

var procsFindBuildPpid = `
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
WHERE proc_build_id = $1
  AND proc_ppid = $2
  AND proc_name = $3
`

var procsDeleteBuild = `
DELETE FROM procs WHERE proc_build_id = $1
`

var registryFindRepo = `
SELECT
 registry_id
,registry_repo_id
,registry_addr
,registry_username
,registry_password
,registry_email
,registry_token
FROM registry
WHERE registry_repo_id = $1
`

var registryFindRepoAddr = `
SELECT
 registry_id
,registry_repo_id
,registry_addr
,registry_username
,registry_password
,registry_email
,registry_token
FROM registry
WHERE registry_repo_id = $1
  AND registry_addr = $2
`

var registryDeleteRepo = `
DELETE FROM registry WHERE registry_repo_id = $1
`

var registryDelete = `
DELETE FROM registry WHERE registry_id = $1
`

var repoUpdateCounter = `
UPDATE repos SET repo_counter = $1
WHERE repo_counter = $2
  AND repo_id = $3
`

var repoFindUser = `
SELECT
 repo_id
,repo_user_id
,repo_owner
,repo_name
,repo_full_name
,repo_avatar
,repo_link
,repo_clone
,repo_branch
,repo_timeout
,repo_private
,repo_trusted
,repo_active
,repo_allow_pr
,repo_hash
,repo_scm
,repo_config_path
,repo_gated
,repo_visibility
,repo_counter
FROM repos
INNER JOIN perms ON perms.perm_repo_id = repos.repo_id
WHERE perms.perm_user_id = $1
ORDER BY repo_full_name ASC
`

var repoInsertIgnore = `
INSERT INTO repos (
 repo_user_id
,repo_owner
,repo_name
,repo_full_name
,repo_avatar
,repo_link
,repo_clone
,repo_branch
,repo_timeout
,repo_private
,repo_trusted
,repo_active
,repo_allow_pr
,repo_hash
,repo_scm
,repo_config_path
,repo_gated
,repo_visibility
,repo_counter
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
ON CONFLICT (repo_full_name) DO NOTHING
`

var repoDelete = `
DELETE FROM repos
WHERE repo_id = $1
`

var secretFindRepo = `
SELECT
 secret_id
,secret_repo_id
,secret_name
,secret_value
,secret_images
,secret_events
,secret_conceal
,secret_skip_verify
FROM secrets
WHERE secret_repo_id = $1
`

var secretFindRepoName = `
SELECT
secret_id
,secret_repo_id
,secret_name
,secret_value
,secret_images
,secret_events
,secret_conceal
,secret_skip_verify
FROM secrets
WHERE secret_repo_id = $1
  AND secret_name = $2
`

var secretDelete = `
DELETE FROM secrets WHERE secret_id = $1
`

var senderFindRepo = `
SELECT
 sender_id
,sender_repo_id
,sender_login
,sender_allow
,sender_block
FROM senders
WHERE sender_repo_id = $1
`

var senderFindRepoLogin = `
SELECT
 sender_id
,sender_repo_id
,sender_login
,sender_allow
,sender_block
FROM senders
WHERE sender_repo_id = $1
  AND sender_login = $2
`

var senderDeleteRepo = `
DELETE FROM senders WHERE sender_repo_id = $1
`

var senderDelete = `
DELETE FROM senders WHERE sender_id = $1
`

var taskList = `
SELECT
 task_id
,task_data
,task_labels
,task_dependencies
,task_run_on
FROM tasks
`

var taskDelete = `
DELETE FROM tasks WHERE task_id = $1
`

var userFind = `
SELECT
 user_id
,user_login
,user_token
,user_secret
,user_expiry
,user_email
,user_avatar
,user_active
,user_synced
,user_admin
,user_hash
FROM users
ORDER BY user_login ASC
`

var userFindLogin = `
SELECT
 user_id
,user_login
,user_token
,user_secret
,user_expiry
,user_email
,user_avatar
,user_active
,user_synced
,user_admin
,user_hash
FROM users
WHERE user_login = $1
LIMIT 1
`

var userUpdate = `
UPDATE users
SET
,user_token  = $1
,user_secret = $2
,user_expiry = $3
,user_email  = $4
,user_avatar = $5
,user_active = $6
,user_synced = $7
,user_admin  = $8
,user_hash   = $9
WHERE user_id = $10
`

var userDelete = `
DELETE FROM users WHERE user_id = $1
`
