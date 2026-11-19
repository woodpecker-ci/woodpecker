package blocks

import "testing"

type TestRepo struct {
}

func NewTestRepo() *TestRepo {
	return &TestRepo{}
}

func (r *TestRepo) Enable(t *testing.T) {

}

func (r *TestRepo) Repair(t *testing.T) {

}

func (r *TestRepo) Disable(t *testing.T) {

}

func (r *TestRepo) Delete(t *testing.T) {

}
