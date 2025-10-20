// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExitError(t *testing.T) {
	err := ExitError{
		UUID: "14534321",
		Code: 255,
	}
	assert.Equal(t, "uuid=14534321: exit code 255", err.Error())
}

func TestOomError(t *testing.T) {
	err := OomError{
		UUID: "14534321",
	}
	assert.Equal(t, "uuid=14534321: received oom kill", err.Error())
}
