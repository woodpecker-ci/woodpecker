-- name: add-builds-changed_files-column
ALTER TABLE builds ADD COLUMN changed_files TEXT;

-- name: update-builds-set-changed_files
UPDATE builds SET changed_files='[]'