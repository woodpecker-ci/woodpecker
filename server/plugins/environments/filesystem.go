package environments

import (
	"strings"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type builtin struct {
	globals []*model.Environ
}

// Filesystem returns a new local registry service.
func Filesystem(params []string) model.EnvironService {
	var globals []*model.Environ

	for _, item := range params {
		kvPair := strings.SplitN(item, ":", 2)
		if len(kvPair) != 2 {
			// ignore items only containing a key and no value
			continue
		}
		globals = append(globals, &model.Environ{Name: kvPair[0], Value: kvPair[1]})
	}
	return &builtin{globals}
}

func (b *builtin) EnvironList(repo *model.Repo) ([]*model.Environ, error) {
	return b.globals, nil
}
