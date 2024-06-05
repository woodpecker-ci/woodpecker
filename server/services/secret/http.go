package secret

import (
	"context"
	"crypto"
	"fmt"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/utils"
)

type http struct {
	Service
	endpoint   string
	privateKey crypto.PrivateKey
}

// New returns a new local secret service.
func NewHTTP(parent Service, endpoint string, privateKey crypto.PrivateKey) Service {
	return &http{parent, endpoint, privateKey}
}

func (h *http) SecretList(ctx context.Context, repo *model.Repo, p *model.ListOptions) (secrets []*model.Secret, err error) {
	path := fmt.Sprintf("%s/repo/%d/secrets?%s", h.endpoint, repo.ID, p.Encode())
	_, err = utils.Send(ctx, "GET", path, h.privateKey, nil, &secrets)
	if err != nil {
		log.Debug().Err(err).Int64("repo-id", repo.ID).Msg("failed to list secrets")
		return nil, err
	}

	return secrets, nil
}

func (h *http) SecretFind(ctx context.Context, repo *model.Repo, name string) (secret *model.Secret, err error) {
	path := fmt.Sprintf("%s/repo/%d/secrets/%s", h.endpoint, repo.ID, name)
	_, err = utils.Send(ctx, "GET", path, h.privateKey, nil, secret)
	if err != nil {
		log.Debug().Err(err).Int64("repo-id", repo.ID).Msgf("failed to get secret '%s'", name)
		return nil, err
	}

	return secret, nil
}

func (h *http) SecretCreate(ctx context.Context, repo *model.Repo, in *model.Secret) (err error) {
	path := fmt.Sprintf("%s/repo/%d/secrets", h.endpoint, repo.ID)
	_, err = utils.Send(ctx, "POST", path, h.privateKey, in, nil)
	if err != nil {
		log.Debug().Err(err).Int64("repo-id", repo.ID).Msgf("failed to create secret")
		return err
	}

	return nil
}

func (h *http) SecretUpdate(ctx context.Context, repo *model.Repo, in *model.Secret) (err error) {
	path := fmt.Sprintf("%s/repo/%d/secrets/%s", h.endpoint, repo.ID, repo.Name)
	_, err = utils.Send(ctx, "PUT", path, h.privateKey, in, nil)
	if err != nil {
		log.Debug().Err(err).Int64("repo-id", repo.ID).Msgf("failed to update secret '%s'", in.Name)
		return err
	}

	return nil
}

func (h *http) SecretDelete(ctx context.Context, repo *model.Repo, name string) (err error) {
	path := fmt.Sprintf("%s/repo/%d/secrets/%s", h.endpoint, repo.ID, name)
	_, err = utils.Send(ctx, "DELETE", path, h.privateKey, nil, nil)
	if err != nil {
		log.Debug().Err(err).Int64("repo-id", repo.ID).Msgf("failed to delete secret '%s'", name)
		return err
	}

	return nil
}
