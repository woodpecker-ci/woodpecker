package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/graph/generated"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (r *buildResolver) Event(ctx context.Context, obj *model.Build) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) Status(ctx context.Context, obj *model.Build) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Build returns generated.BuildResolver implementation.
func (r *Resolver) Build() generated.BuildResolver { return &buildResolver{r} }

type buildResolver struct{ *Resolver }
