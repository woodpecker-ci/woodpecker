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

package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
)

func TestNewTaskTimeoutValidation(t *testing.T) {
	t.Run("zero falls back to default", func(t *testing.T) {
		q, err := New(t.Context(), Config{Backend: TypeMemory})
		require.NoError(t, err)
		require.NotNil(t, q)
	})

	t.Run("custom timeout is accepted", func(t *testing.T) {
		q, err := New(t.Context(), Config{Backend: TypeMemory, TaskTimeout: 15 * time.Minute})
		require.NoError(t, err)
		require.NotNil(t, q)
	})

	t.Run("below default is rejected", func(t *testing.T) {
		_, err := New(t.Context(), Config{Backend: TypeMemory, TaskTimeout: 30 * time.Second})
		assert.ErrorContains(t, err, constant.TaskTimeout.String())
	})
}
