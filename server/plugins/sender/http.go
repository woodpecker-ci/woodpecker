package sender

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/utils"
)

type http struct {
	endpoint string
	secret   string
}

// NewRemote returns a new remote gating service.
func NewHTTP(endpoint, secret string) model.SenderService {
	return &http{endpoint, secret}
}

func (p *http) SenderAllowed(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build, conf *model.Config) (bool, error) {
	path := fmt.Sprintf("%s/senders/%s/%s/%s/verify", p.endpoint, repo.Owner, repo.Name, build.Sender)
	data := map[string]interface{}{
		"build":  build,
		"config": conf,
	}
	_, err := utils.Send(ctx, "POST", path, p.secret, &data, nil)
	if err != nil {
		return false, err
	}
	return true, err
}

func (p *http) SenderCreate(ctx context.Context, repo *model.Repo, sender *model.Sender) error {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	_, err := utils.Send(ctx, "POST", path, p.secret, sender, nil)
	return err
}

func (p *http) SenderUpdate(ctx context.Context, repo *model.Repo, sender *model.Sender) error {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	_, err := utils.Send(ctx, "PUT", path, p.secret, sender, nil)
	return err
}

func (p *http) SenderDelete(ctx context.Context, repo *model.Repo, login string) error {
	path := fmt.Sprintf("%s/senders/%s/%s/%s", p.endpoint, repo.Owner, repo.Name, login)
	_, err := utils.Send(ctx, "DELETE", path, p.secret, nil, nil)
	return err
}

func (p *http) SenderList(ctx context.Context, repo *model.Repo) (out []*model.Sender, err error) {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "GET", path, p.secret, nil, out)
	return out, err
}
