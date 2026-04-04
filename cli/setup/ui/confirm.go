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
	"charm.land/huh/v2"
)

func Confirm(prompt string) (bool, error) {
	var confirm bool
	err := huh.NewConfirm().
		Title(prompt).
		Affirmative("Yes!").
		Negative("No.").
		Value(&confirm).Run()
	if err != nil {
		return false, err
	}

	return confirm, err
}
