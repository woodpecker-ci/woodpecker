package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/graph/generated"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (r *secretResolver) Events(ctx context.Context, obj *model.Secret) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Secret returns generated.SecretResolver implementation.
func (r *Resolver) Secret() generated.SecretResolver { return &secretResolver{r} }

type secretResolver struct{ *Resolver }
