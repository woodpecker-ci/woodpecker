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

package compiler

import "fmt"

type ErrExtraHostFormat struct {
	host string
}

func (err *ErrExtraHostFormat) Error() string {
	return fmt.Sprintf("extra host %s is in wrong format", err.host)
}

func (*ErrExtraHostFormat) Is(target error) bool {
	_, ok := target.(*ErrExtraHostFormat) //nolint:errorlint
	return ok
}

type ErrStepDependencyCycle struct {
	path []string
}

func (err *ErrStepDependencyCycle) Error() string {
	return fmt.Sprintf("cycle detected: %v", err.path)
}

func (*ErrStepDependencyCycle) Is(target error) bool {
	_, ok := target.(*ErrStepDependencyCycle) //nolint:errorlint
	return ok
}
