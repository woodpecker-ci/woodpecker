// Copyright 2024 Woodpecker Authors
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
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
)

var (
	bold      = lipgloss.NewStyle().Bold(true)
	underline = lipgloss.NewStyle().Underline(true)
)

// FormatLintError returns the issues of err as a tree below the file name and
// an error naming the file if one of them is an error.
func FormatLintError(file string, err error, strict bool) (string, error) {
	if err == nil {
		return "", nil
	}

	issues := tree.New()
	amountErrors := 0
	amountWarnings := 0
	linterErrors := pipeline_errors.GetPipelineErrors(err)
	for _, err := range linterErrors {
		icon := "❌"
		if !strict && err.IsWarning {
			icon = "⚠️ "
			amountWarnings++
		} else {
			amountErrors++
		}

		if data := pipeline_errors.GetLinterData(err); data != nil {
			issues.Child(fmt.Sprintf("%s %s  %s", icon, bold.Render(data.Field), err.Message))
		} else {
			issues.Child(fmt.Sprintf("%s %s", icon, err.Message))
		}
	}

	if amountErrors > 0 {
		root := fmt.Sprintf("🔥 %s has %d errors:", underline.Render(file), len(linterErrors))
		if amountWarnings > 0 {
			root = fmt.Sprintf("🔥 %s has %d errors and warnings:", underline.Render(file), len(linterErrors))
		}
		return issues.Root(root).String() + "\n", fmt.Errorf("%s has errors", file)
	}

	return issues.Root(fmt.Sprintf("⚠️  %s has %d warnings:", underline.Render(file), len(linterErrors))).String() + "\n", nil
}
