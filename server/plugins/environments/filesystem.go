package environments

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
	"strings"
)

type builtin struct {
	globals []*model.Environ
}

// New returns a new local registry service.
func Filesystem(params []string) model.EnvironService {
	var globals []*model.Environ

	for _, item := range params {
		kvpair := strings.SplitN(item, ":", 2)
		globals = append(globals, &model.Environ{Name: kvpair[0], Value: kvpair[1]})
	}
	return &builtin{globals}
}

func (b *builtin) EnvironList(repo *model.Repo) ([]*model.Environ, error) {
	return b.globals, nil
}
