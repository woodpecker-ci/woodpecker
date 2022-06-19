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

func (r *buildResolver) CreatedAt(ctx context.Context, obj *model.Build) (int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) UpdatedAt(ctx context.Context, obj *model.Build) (int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) EnqueuedAt(ctx context.Context, obj *model.Build) (int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) StartedAt(ctx context.Context, obj *model.Build) (int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) FinishedAt(ctx context.Context, obj *model.Build) (int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) DeployTo(ctx context.Context, obj *model.Build) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) AuthorAvatar(ctx context.Context, obj *model.Build) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) AuthorEmail(ctx context.Context, obj *model.Build) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) LinkURL(ctx context.Context, obj *model.Build) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) ReviewedBy(ctx context.Context, obj *model.Build) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *buildResolver) ReviewedAt(ctx context.Context, obj *model.Build) (int, error) {
	panic(fmt.Errorf("not implemented"))
}

// Build returns generated.BuildResolver implementation.
func (r *Resolver) Build() generated.BuildResolver { return &buildResolver{r} }

type buildResolver struct{ *Resolver }
