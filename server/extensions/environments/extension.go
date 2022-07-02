package environments

import "github.com/woodpecker-ci/woodpecker/server/model"

// EnvironExtension defines an extension for managing environment variables.
type EnvironExtension interface {
	EnvironList(*model.Repo) ([]*model.Environ, error)
}
