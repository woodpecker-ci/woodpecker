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

package registry

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

// Service defines a service for managing registries.
type Service interface {
	RegistryFind(*model.Repo, string) (*model.Registry, error)
	RegistryList(*model.Repo, *model.ListOptions) ([]*model.Registry, error)
	RegistryCreate(*model.Repo, *model.Registry) error
	RegistryUpdate(*model.Repo, *model.Registry) error
	RegistryDelete(*model.Repo, string) error
}

// ReadOnlyService defines a service for managing registries.
type ReadOnlyService interface {
	RegistryFind(*model.Repo, string) (*model.Registry, error)
	RegistryList(*model.Repo, *model.ListOptions) ([]*model.Registry, error)
}
