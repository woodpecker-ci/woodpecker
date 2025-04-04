package utils

import "testing"

type TestRepo struct {
	folder string
}

func (r *TestRepo) Clone(t *testing.T, sourcePath, remoteURL string) error {
	r.folder = t.TempDir()

	NewTask("cp", "-r", sourcePath, r.folder).RunOrFail(t)
	NewTask("git", "init").RunOrFail(t)
	NewTask("git", "remote", "add", remoteURL).RunOrFail(t)

	r.Commit(t, ":tada: init")

	r.Push(t)

	return nil
}

func (r *TestRepo) Commit(t *testing.T, message string) {
	NewTask("git", "commit", "-m", message).RunOrFail(t)
}

func (r *TestRepo) Push(t *testing.T) {
	NewTask("git", "push", "-u", "origin", "main").RunOrFail(t)
}

func (r *TestRepo) Tag(t *testing.T, name, message string) {
	NewTask("git", "tag", "-a", name, "-m", message).RunOrFail(t)
}
