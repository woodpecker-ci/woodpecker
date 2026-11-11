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

package lint

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
)

var ansiCodes = regexp.MustCompile("\x1b\\[[0-9;]*m")

func TestFormatLintError(t *testing.T) {
	warning := &pipeline_errors.PipelineError{
		Type:      pipeline_errors.PipelineErrorTypeLinter,
		Message:   "Consider adding a `when` block",
		IsWarning: true,
		Data:      &pipeline_errors.LinterErrorData{File: "build", Field: "steps.build"},
	}

	t.Run("errors", func(t *testing.T) {
		str, err := FormatLintError("build.yaml", multierr.Combine(warning, errors.New("yaml: line 3: broken")), false)
		assert.EqualError(t, err, "build.yaml has errors")
		assert.Equal(t, "🔥 build.yaml has 2 errors and warnings:\n"+
			"├── ⚠️  steps.build  Consider adding a `when` block\n"+
			"└── ❌ yaml: line 3: broken\n", ansiCodes.ReplaceAllString(str, ""))
	})

	t.Run("warnings", func(t *testing.T) {
		str, err := FormatLintError("build.yaml", warning, false)
		assert.NoError(t, err)
		assert.Equal(t, "⚠️  build.yaml has 1 warnings:\n"+
			"└── ⚠️  steps.build  Consider adding a `when` block\n", ansiCodes.ReplaceAllString(str, ""))
	})

	t.Run("strict", func(t *testing.T) {
		_, err := FormatLintError("build.yaml", warning, true)
		assert.EqualError(t, err, "build.yaml has errors")
	})
}
