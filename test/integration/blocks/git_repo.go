package blocks

import (
	"os"
	"path"
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

type gitRepo struct {
	folder string
}

func NewGitRepo(t *testing.T) *gitRepo {
	return &gitRepo{
		folder: t.TempDir(),
	}
}

func (r *gitRepo) Clone(t *testing.T, remoteURL string) {
	utils.NewCommand("git", "clone", remoteURL, r.folder).RunOrFail(t)
}

func (r *gitRepo) Init(t *testing.T, remoteURL string) {
	utils.NewCommand("git", "init").RunOrFail(t)
	utils.NewCommand("git", "remote", "add", remoteURL).RunOrFail(t)
	utils.NewCommand("git", "branch", "set-upstream-to=origin/main").RunOrFail(t)
}

func (r *gitRepo) InitFromTemplate(t *testing.T, templatePath, remoteURL string) {
	utils.NewCommand("cp", "-r", templatePath, r.folder).RunOrFail(t)
	r.Init(t, remoteURL)
}

func (r *gitRepo) Add(t *testing.T, filePath string) {
	utils.NewCommand("git", "add", filePath).RunOrFail(t)
}

func (r *gitRepo) Commit(t *testing.T, message string) {
	utils.NewCommand("git", "commit", "-m", message).RunOrFail(t)
}

func (r *gitRepo) Push(t *testing.T) {
	utils.NewCommand("git", "push").RunOrFail(t)
}

func (r *gitRepo) Tag(t *testing.T, name, message string) {
	utils.NewCommand("git", "tag", "-a", name, "-m", message).RunOrFail(t)
}

func (r *gitRepo) WriteFile(t *testing.T, filePath string, content []byte) {
	// Ensure the directory exists
	os.MkdirAll(path.Join(r.folder, path.Dir(filePath)), 0755)

	err := os.WriteFile(path.Join(r.folder, filePath), content, 0644)
	if err != nil {
		t.Fatalf("Failed to write file %s: %v", filePath, err)
	}
}
