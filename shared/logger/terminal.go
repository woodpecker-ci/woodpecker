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

package logger

import (
	"os"

	"golang.org/x/term"
)

// IsInteractiveTerminal reports whether stdout is attached to a terminal.
// It is the single source of truth for this check across the codebase;
// do not re-implement it.
func IsInteractiveTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
