package config

import (
	"context"
	"crypto"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/extensions/utils"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

type HttpFetcher struct {
	endpoint   string
	privateKey crypto.PrivateKey
}

// Same as remote.FileMeta but with json tags and string data
type httpConfig struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type httpConfigRequestStructure struct {
	Repo          *model.Repo   `json:"repo"`
	Build         *model.Build  `json:"build"`
	Configuration []*httpConfig `json:"configs"`
}

type httpResponseStructure struct {
	Configs []httpConfig `json:"configs"`
}

func NewHTTP(endpoint string, privateKey crypto.PrivateKey) *HttpFetcher {
	return &HttpFetcher{endpoint, privateKey}
}

func (cp *HttpFetcher) FetchConfig(ctx context.Context, _ *model.User, repo *model.Repo, build *model.Build, currentFileMeta []*remote.FileMeta) (configData []*remote.FileMeta, useOld bool, err error) {
	currentConfigs := make([]*httpConfig, len(currentFileMeta))
	for i, pipe := range currentFileMeta {
		currentConfigs[i] = &httpConfig{Name: pipe.Name, Data: string(pipe.Data)}
	}

	response := new(httpResponseStructure)
	body := httpConfigRequestStructure{Repo: repo, Build: build, Configuration: currentConfigs}
	status, err := utils.Send(ctx, "POST", cp.endpoint, cp.privateKey, body, response)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to fetch config via http (%d) %w", status, err)
	}

	var newFileMeta []*remote.FileMeta
	if status != 200 {
		newFileMeta = make([]*remote.FileMeta, 0)
	} else {
		newFileMeta = make([]*remote.FileMeta, len(response.Configs))
		for i, pipe := range response.Configs {
			newFileMeta[i] = &remote.FileMeta{Name: pipe.Name, Data: []byte(pipe.Data)}
		}
	}

	return newFileMeta, status == 204, nil
}
