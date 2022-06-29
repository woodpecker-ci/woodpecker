package config

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/extensions/utils"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

type http struct {
	endpoint   string
	privateKey crypto.PrivateKey
}

// Same as remote.FileMeta but with json tags and string data
type config struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type requestStructure struct {
	Repo          *model.Repo  `json:"repo"`
	Build         *model.Build `json:"build"`
	Configuration []*config    `json:"configs"`
}

type responseStructure struct {
	Configs []config `json:"configs"`
}

func NewHTTP(endpoint string, privateKey crypto.PrivateKey) Extension {
	return &http{endpoint, privateKey}
}

func (cp *http) IsConfigured() bool {
	return cp.endpoint != ""
}

func FromRepo(repo *model.Repo) Extension {
	if repo.ConfigEndpoint == "" {
		return nil
	}

	// TODO: create & use global server key
	_, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	return NewHTTP(repo.ConfigEndpoint, privEd25519Key)
}

func (cp *http) FetchConfig(ctx context.Context, repo *model.Repo, build *model.Build, currentFileMeta []*remote.FileMeta) (configData []*remote.FileMeta, useOld bool, err error) {
	currentConfigs := make([]*config, len(currentFileMeta))
	for i, pipe := range currentFileMeta {
		currentConfigs[i] = &config{Name: pipe.Name, Data: string(pipe.Data)}
	}

	response := new(responseStructure)
	body := requestStructure{Repo: repo, Build: build, Configuration: currentConfigs}
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
