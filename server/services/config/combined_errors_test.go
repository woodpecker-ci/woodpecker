// Copyright 2026 Woodpecker Authors
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

package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
)

// fetchFunc is a config service that answers with whatever the test needs.
type fetchFunc func(old []*forge_types.FileMeta) ([]*forge_types.FileMeta, error)

func (f fetchFunc) Fetch(_ context.Context, _ forge.Forge, _ *model.User, _ *model.Repo, _ *model.Pipeline, old []*forge_types.FileMeta, _ bool) ([]*forge_types.FileMeta, error) {
	return f(old)
}

func TestCombinedErrorHandling(t *testing.T) {
	central := []*forge_types.FileMeta{{Name: "central", Data: []byte("steps: []")}}
	notFound := fetchFunc(func([]*forge_types.FileMeta) ([]*forge_types.FileMeta, error) {
		return nil, &forge_types.ErrConfigNotFound{Configs: []string{".woodpecker.yaml"}}
	})
	keep := fetchFunc(func(old []*forge_types.FileMeta) ([]*forge_types.FileMeta, error) { return old, nil })
	fetch := func(services ...config.Service) ([]*forge_types.FileMeta, error) {
		return config.NewCombined(services...).Fetch(t.Context(), nil, nil, &model.Repo{}, &model.Pipeline{}, nil, false)
	}

	t.Run("a failure is not covered up by the next service", func(t *testing.T) {
		forgeDown := errors.New("forge is down")
		asked := false

		files, err := fetch(
			fetchFunc(func([]*forge_types.FileMeta) ([]*forge_types.FileMeta, error) { return nil, forgeDown }),
			fetchFunc(func(old []*forge_types.FileMeta) ([]*forge_types.FileMeta, error) {
				asked = true
				return old, nil
			}),
		)
		require.ErrorIs(t, err, forgeDown)
		assert.Nil(t, files)
		assert.False(t, asked, "the next service must not be asked to continue after a failure")
	})

	t.Run("a repo without config can get one from the next service", func(t *testing.T) {
		files, err := fetch(notFound, fetchFunc(func([]*forge_types.FileMeta) ([]*forge_types.FileMeta, error) { return central, nil }))
		require.NoError(t, err)
		assert.Equal(t, central, files)
	})

	t.Run("a repo without config stays without if no service provides one", func(t *testing.T) {
		files, err := fetch(notFound, keep)
		require.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})
		assert.Empty(t, files)
	})
}
