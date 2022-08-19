package datastore

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
	"xorm.io/xorm"
)

func (s storage) GetRedirection(fullName string) (*model.Redirection, error) {
	repo := new(model.Redirection)
	return repo, wrapGet(s.engine.Where("repo_full_name = ?", fullName).Get(repo))
}

func (s storage) CreateRedirection(redirect *model.Redirection) error {
	// only Insert set auto created ID back to object
	sess := s.engine.NewSession()
	defer sess.Close()
	return s.createRedirection(sess, redirect)
}

func (s storage) createRedirection(e *xorm.Session, redirect *model.Redirection) error {
	// only Insert set auto created ID back to object
	_, err := e.Insert(redirect)
	return err
}

func (s storage) HasRedirectionForRepo(repoID int64, fullName string) (bool, error) {
	return s.engine.Where("repo_id = ? AND repo_full_name = ?", repoID, fullName).Get(new(model.Redirection))
}
