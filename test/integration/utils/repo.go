package utils

import "testing"

type TestRepo struct {
	folder string
}

func (r *TestRepo) Clone(t *testing.T, sourcePath, remoteURL string) error {
	r.folder = t.TempDir()

	runOrFail(t, "cp", "-r", sourcePath, r.folder)

	runOrFail(t, "git", "init")

	runOrFail(t, "git", "remote", "add", remoteURL)

	r.Commit(t, ":tada: init")

	r.Push(t)

	return nil
}

func (r *TestRepo) Commit(t *testing.T, message string) {
	runOrFail(t, "git", "commit", "-m", message)
}

func (r *TestRepo) Push(t *testing.T) {
	runOrFail(t, "git", "push", "-u", "origin", "main")
}

func (r *TestRepo) Tag(t *testing.T, name, message string) {
	runOrFail(t, "git", "tag", "-a", name, "-m", message)
}
