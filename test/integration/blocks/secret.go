package blocks

import "testing"

type TestSecret struct {
	repo *TestRepo
}

func NewTestSecret(repo *TestRepo) *TestSecret {
	return &TestSecret{repo: repo}
}

func (s *TestSecret) Create(t *testing.T, key, value string) {

}

func (s *TestSecret) Update(t *testing.T, value string) {

}

func (s *TestSecret) Delete(t *testing.T) {

}
