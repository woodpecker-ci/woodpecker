package environments

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

// Service defines a service for managing environment variables.
type Service interface {
	EnvironList(*model.Repo) ([]*model.Environ, error)
}
