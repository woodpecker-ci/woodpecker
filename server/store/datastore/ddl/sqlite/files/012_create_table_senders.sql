-- name: create-table-senders

CREATE TABLE IF NOT EXISTS senders (
 sender_id      INTEGER PRIMARY KEY AUTOINCREMENT
,sender_repo_id INTEGER
,sender_login   TEXT
,sender_allow   BOOLEAN
,sender_block   BOOLEAN

,UNIQUE(sender_repo_id,sender_login)
);

-- name: create-index-sender-repos

CREATE INDEX IF NOT EXISTS sender_repo_ix ON senders (sender_repo_id);
