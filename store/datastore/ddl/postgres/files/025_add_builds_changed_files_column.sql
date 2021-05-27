-- name: add-builds-changed_files-column
ALTER TABLE builds ADD COLUMN changed_files TEXT;