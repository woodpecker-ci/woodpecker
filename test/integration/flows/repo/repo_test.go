package repo_test

import (
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/blocks"
)

func TestFlow_EnableRepo(t *testing.T) {
	t.Parallel()

	repo := blocks.NewTestRepo()
	repo.Enable(t)
}

func TestFlow_RepairRepo(t *testing.T) {
	t.Parallel()

	repo := blocks.NewTestRepo()
	repo.Enable(t)
	repo.Repair(t)
}

func TestFlow_DisableRepo(t *testing.T) {
	t.Parallel()

	repo := blocks.NewTestRepo()
	repo.Enable(t)
	repo.Disable(t)
}

func TestFlow_DeleteRepo(t *testing.T) {
	t.Parallel()

	repo := blocks.NewTestRepo()
	repo.Enable(t)
	repo.Delete(t)
}
