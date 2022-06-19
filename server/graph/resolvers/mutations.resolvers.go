package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/graph/generated"
	"github.com/woodpecker-ci/woodpecker/server/graph/model"
	model1 "github.com/woodpecker-ci/woodpecker/server/model"
)

func (r *mutationResolver) ActivateRepo(ctx context.Context, owner string, name string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateRepo(ctx context.Context, input model.UpdateRepo) (*model1.Repo, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteRepo(ctx context.Context, owner string, name string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RepairRepo(ctx context.Context, owner string, name string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CancelBuild(ctx context.Context, owner string, name string, buildID int) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ApproveBuild(ctx context.Context, owner string, name string, buildID int) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeclineBuild(ctx context.Context, owner string, name string, buildID int) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RestartBuild(ctx context.Context, owner string, name string, buildID int) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateSecret(ctx context.Context, owner string, name string, secret model.NewSecret) (*model1.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateSecret(ctx context.Context, owner string, name string, secret model.UpdateSecret) (*model1.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteSecret(ctx context.Context, owner string, name string, secretName string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateRegistry(ctx context.Context, owner string, name string, registry model.NewRegistry) (*model1.Registry, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateRegistry(ctx context.Context, owner string, name string, registry model.UpdateRegistry) (*model1.Registry, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteRegistry(ctx context.Context, owner string, name string, registryName string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
