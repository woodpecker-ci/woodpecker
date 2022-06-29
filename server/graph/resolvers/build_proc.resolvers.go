package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/graph/generated"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (r *procResolver) State(ctx context.Context, obj *model.Proc) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *procResolver) Environ(ctx context.Context, obj *model.Proc) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Proc returns generated.ProcResolver implementation.
func (r *Resolver) Proc() generated.ProcResolver { return &procResolver{r} }

type procResolver struct{ *Resolver }
