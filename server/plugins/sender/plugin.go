package sender

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/internal"
)

type plugin struct {
	endpoint string
}

// NewRemote returns a new remote gating service.
func NewRemote(endpoint string) model.SenderService {
	return &plugin{endpoint}
}

func (p *plugin) SenderAllowed(user *model.User, repo *model.Repo, build *model.Build, conf *model.Config) (bool, error) {
	path := fmt.Sprintf("%s/senders/%s/%s/%s/verify", p.endpoint, repo.Owner, repo.Name, build.Sender)
	data := map[string]interface{}{
		"build":  build,
		"config": conf,
	}
	err := internal.Send(context.TODO(), "POST", path, &data, nil)
	if err != nil {
		return false, err
	}
	return true, err
}

func (p *plugin) SenderCreate(repo *model.Repo, sender *model.Sender) error {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	return internal.Send(context.TODO(), "POST", path, sender, nil)
}

func (p *plugin) SenderUpdate(repo *model.Repo, sender *model.Sender) error {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	return internal.Send(context.TODO(), "PUT", path, sender, nil)
}

func (p *plugin) SenderDelete(repo *model.Repo, login string) error {
	path := fmt.Sprintf("%s/senders/%s/%s/%s", p.endpoint, repo.Owner, repo.Name, login)
	return internal.Send(context.TODO(), "DELETE", path, nil, nil)
}

func (p *plugin) SenderList(repo *model.Repo) (out []*model.Sender, err error) {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	err = internal.Send(context.TODO(), "GET", path, nil, out)
	return out, err
}
