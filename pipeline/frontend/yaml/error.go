// Copyright 2022 Woodpecker Authors
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

package yaml

import "errors"

var (
	ErrUnsuportedVersion = errors.New("unsuported yaml version detected")
	ErrMissingVersion    = errors.New("missing yaml version")
)

// PipelineParseError is an error that occurs when the pipeline parsing fails.
type PipelineParseError struct {
	Err error
}

func (e PipelineParseError) Error() string {
	return e.Err.Error()
}

func (e PipelineParseError) Is(err error) bool {
	target1 := PipelineParseError{}
	target2 := &target1
	return errors.As(err, &target1) || errors.As(err, &target2)
}
