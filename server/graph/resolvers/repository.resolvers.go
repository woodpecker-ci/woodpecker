package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/graph/generated"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (r *repositoryResolver) SCMKind(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repositoryResolver) Visibility(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repositoryResolver) CancelPreviousPipelineEvents(ctx context.Context, obj *model.Repo) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repositoryResolver) LastBuild(ctx context.Context, obj *model.Repo) (int, error) {
	panic(fmt.Errorf("not implemented"))
}

// Repository returns generated.RepositoryResolver implementation.
func (r *Resolver) Repository() generated.RepositoryResolver { return &repositoryResolver{r} }

type repositoryResolver struct{ *Resolver }
