// Copyright 2023 Woodpecker Authors
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

package datastore

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

func TestWrapGet(t *testing.T) {
	err := wrapGet(false, nil)
	assert.ErrorIs(t, err, types.RecordNotExist)

	err = wrapGet(true, errors.New("test err"))
	assert.Equal(t, "TestWrapGet: test err", err.Error())
}

func TestWrapDelete(t *testing.T) {
	err := wrapDelete(0, nil)
	assert.ErrorIs(t, err, types.RecordNotExist)

	err = wrapDelete(1, errors.New("test err"))
	assert.Equal(t, "TestWrapDelete: test err", err.Error())
}
