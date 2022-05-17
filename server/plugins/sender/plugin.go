package sender

import (
	"context"
	"crypto"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/utils"
)

type plugin struct {
	endpoint   string
	privateKey crypto.PrivateKey
}

// NewRemote returns a new remote gating service.
func NewRemote(endpoint string, privateKey crypto.PrivateKey) model.SenderService {
	return &plugin{endpoint, privateKey}
}

func (p *plugin) SenderAllowed(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build, conf *model.Config) (bool, error) {
	path := fmt.Sprintf("%s/senders/%s/%s/%s/verify", p.endpoint, repo.Owner, repo.Name, build.Sender)
	data := map[string]interface{}{
		"build":  build,
		"config": conf,
	}
	_, err := utils.Send(ctx, "POST", path, p.privateKey, &data, nil)
	if err != nil {
		return false, err
	}
	return true, err
}

func (p *plugin) SenderCreate(ctx context.Context, repo *model.Repo, sender *model.Sender) error {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	_, err := utils.Send(ctx, "POST", path, p.privateKey, sender, nil)
	return err
}

func (p *plugin) SenderUpdate(ctx context.Context, repo *model.Repo, sender *model.Sender) error {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	_, err := utils.Send(ctx, "PUT", path, p.privateKey, sender, nil)
	return err
}

func (p *plugin) SenderDelete(ctx context.Context, repo *model.Repo, login string) error {
	path := fmt.Sprintf("%s/senders/%s/%s/%s", p.endpoint, repo.Owner, repo.Name, login)
	_, err := utils.Send(ctx, "DELETE", path, p.privateKey, nil, nil)
	return err
}

func (p *plugin) SenderList(ctx context.Context, repo *model.Repo) (out []*model.Sender, err error) {
	path := fmt.Sprintf("%s/senders/%s/%s", p.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "GET", path, p.privateKey, nil, out)
	return out, err
}
