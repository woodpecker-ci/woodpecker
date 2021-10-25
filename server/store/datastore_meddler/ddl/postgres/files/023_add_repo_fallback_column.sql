-- name: alter-table-add-repo-fallback
ALTER TABLE repos ADD COLUMN repo_fallback BOOLEAN

-- name: update-table-set-repo-fallback
UPDATE repos SET repo_fallback='false'
