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

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

//go:generate mockery --name Service --output mocks --case underscore

// Service defines a service for managing secrets.
type Service interface {
	SecretListPipeline(*model.Repo, *model.Pipeline) ([]*model.Secret, error)
	// Repository secrets
	SecretFind(*model.Repo, string) (*model.Secret, error)
	SecretList(*model.Repo, *model.ListOptions) ([]*model.Secret, error)
	SecretCreate(*model.Repo, *model.Secret) error
	SecretUpdate(*model.Repo, *model.Secret) error
	SecretDelete(*model.Repo, string) error
	// Organization secrets
	OrgSecretFind(int64, string) (*model.Secret, error)
	OrgSecretList(int64, *model.ListOptions) ([]*model.Secret, error)
	OrgSecretCreate(int64, *model.Secret) error
	OrgSecretUpdate(int64, *model.Secret) error
	OrgSecretDelete(int64, string) error
	// Global secrets
	GlobalSecretFind(string) (*model.Secret, error)
	GlobalSecretList(*model.ListOptions) ([]*model.Secret, error)
	GlobalSecretCreate(*model.Secret) error
	GlobalSecretUpdate(*model.Secret) error
	GlobalSecretDelete(string) error
}
