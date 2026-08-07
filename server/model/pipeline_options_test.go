// Copyright 2026 Woodpecker Authors
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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineOptionsValidate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, (&PipelineOptions{Branch: "main"}).Validate())
	assert.NoError(t, (&PipelineOptions{Tag: "v1.0.0"}).Validate())
	assert.NoError(t, (&PipelineOptions{SHA: "abc1234"}).Validate())

	assert.Error(t, (&PipelineOptions{}).Validate())
	assert.Error(t, (&PipelineOptions{Branch: "main", Tag: "v1"}).Validate())
	assert.Error(t, (&PipelineOptions{Branch: "main", SHA: "abc"}).Validate())
	assert.Error(t, (&PipelineOptions{Tag: "v1", SHA: "abc"}).Validate())
}
