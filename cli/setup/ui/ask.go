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

package ui

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
)

func Ask(prompt, placeholder string, required bool) (string, error) {
	var input string
	err := huh.NewInput().
		Title(prompt).
		Value(&input).
		Placeholder(placeholder).Validate(func(s string) error {
		if required && strings.TrimSpace(s) == "" {
			return errors.New("required")
		}
		return nil
	}).Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}
