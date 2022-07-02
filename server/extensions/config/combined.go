package config

import (
	"context"
	"crypto"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

type combinedFetcher struct {
	globalConfigEndpoint string
	signaturePrivateKey  crypto.PrivateKey
	remote               remote.Remote
}

func NewCombined(remote remote.Remote, globalConfigEndpoint string, signaturePrivateKey crypto.PrivateKey) Extension {
	return &combinedFetcher{
		globalConfigEndpoint: globalConfigEndpoint,
		signaturePrivateKey:  signaturePrivateKey,
		remote:               remote,
	}
}

func (c *combinedFetcher) FetchConfig(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build) (files []*remote.FileMeta, err error) {
	r := newRemote(c.remote)
	files, err = r.FetchConfig(ctx, user, repo, build)
	if err != nil {
		return nil, err
	}

	var configExtension *HttpFetcher
	if repo.ConfigEndpoint != "" {
		configExtension = NewHTTP(repo.ConfigEndpoint, c.signaturePrivateKey)
	} else if c.globalConfigEndpoint != "" {
		configExtension = NewHTTP(c.globalConfigEndpoint, c.signaturePrivateKey)
	}

	if configExtension != nil {
		fetchCtx, cancel := context.WithTimeout(ctx, configFetchTimeout)
		defer cancel() // ok here as we only try http fetching once, returning on fail and success

		log.Trace().Msgf("ConfigFetch[%s]: getting config from external http service", repo.FullName)
		newConfigs, useOld, err := configExtension.FetchConfig(fetchCtx, user, repo, build, files)
		if err != nil {
			log.Error().Msg("Got error " + err.Error())
			return nil, fmt.Errorf("On Fetching config via http : %s", err)
		}

		if !useOld {
			return newConfigs, nil
		}
	}

	return
}
