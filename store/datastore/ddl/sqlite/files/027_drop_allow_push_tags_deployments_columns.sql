-- name: drop-allow-push-tags-deploys-columns
BEGIN TRANSACTION;
CREATE TABLE repos_new (
  repo_id INTEGER PRIMARY KEY AUTOINCREMENT,
  repo_user_id INTEGER,
  repo_owner TEXT,
  repo_name TEXT,
  repo_full_name TEXT,
  repo_avatar TEXT,
  repo_link TEXT,
  repo_clone TEXT,
  repo_branch TEXT,
  repo_timeout INTEGER,
  repo_private BOOLEAN,
  repo_trusted BOOLEAN,
  repo_active BOOLEAN,
  repo_allow_pr BOOLEAN,
  repo_hash TEXT,
  repo_scm TEXT,
  repo_config_path TEXT,
  repo_gated BOOLEAN,
  repo_visibility TEXT,
  repo_counter INTEGER,
  UNIQUE(repo_full_name)
);
INSERT INTO repos_new SELECT repo_id
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
,repo_counter FROM repos;
DROP TABLE repos;
ALTER TABLE repos_new RENAME TO repos;
COMMIT;
