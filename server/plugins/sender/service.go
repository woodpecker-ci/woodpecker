package sender

import "github.com/woodpecker-ci/woodpecker/server/model"

type SenderService interface {
	SenderAllowed(*model.User, *model.Repo, *model.Build, *model.Config) (bool, error)
	SenderCreate(*model.Repo, *model.Sender) error
	SenderUpdate(*model.Repo, *model.Sender) error
	SenderDelete(*model.Repo, string) error
	SenderList(*model.Repo) ([]*model.Sender, error)
}
