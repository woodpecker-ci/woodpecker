// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secret

import (
	"context"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

// Service defines a service for managing secrets.
type Service interface {
	SecretListPipeline(context.Context, *model.Repo, *model.Pipeline, *model.ListOptions) ([]*model.Secret, error)
	// Repository secrets
	SecretFind(context.Context, *model.Repo, string) (*model.Secret, error)
	SecretList(context.Context, *model.Repo, *model.ListOptions) ([]*model.Secret, error)
	SecretCreate(context.Context, *model.Repo, *model.Secret) error
	SecretUpdate(context.Context, *model.Repo, *model.Secret) error
	SecretDelete(context.Context, *model.Repo, string) error
	// Organization secrets
	OrgSecretFind(context.Context, int64, string) (*model.Secret, error)
	OrgSecretList(context.Context, int64, *model.ListOptions) ([]*model.Secret, error)
	OrgSecretCreate(context.Context, int64, *model.Secret) error
	OrgSecretUpdate(context.Context, int64, *model.Secret) error
	OrgSecretDelete(context.Context, int64, string) error
	// Global secrets
	GlobalSecretFind(context.Context, string) (*model.Secret, error)
	GlobalSecretList(context.Context, *model.ListOptions) ([]*model.Secret, error)
	GlobalSecretCreate(context.Context, *model.Secret) error
	GlobalSecretUpdate(context.Context, *model.Secret) error
	GlobalSecretDelete(context.Context, string) error
}
