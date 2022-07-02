package environments

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// TODO: migrate into secrets after renaming secrets to environment
type builtin struct {
	globals []*model.Environ
}

// Parse returns a EnvironService based on a string slice where key and value are separated by a ":" delimiter.
func Parse(params []string) EnvironExtension {
	var globals []*model.Environ

	for _, item := range params {
		kvPair := strings.SplitN(item, ":", 2)
		if len(kvPair) != 2 {
			// ignore items only containing a key and no value
			log.Warn().Msgf("key '%s' has no value, will be ignored", kvPair[0])
			continue
		}
		globals = append(globals, &model.Environ{Name: kvPair[0], Value: kvPair[1]})
	}
	return &builtin{globals}
}

func (b *builtin) EnvironList(repo *model.Repo) ([]*model.Environ, error) {
	return b.globals, nil
}
