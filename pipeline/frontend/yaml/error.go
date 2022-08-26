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

// PipelineParseError is an error that occurs when the pipeline parsing fails.
type PipelineParseError struct {
	Err error
}

func (e PipelineParseError) Error() string {
	return e.Err.Error()
}

func (e PipelineParseError) Is(target error) bool {
	_, ok1 := target.(PipelineParseError)
	_, ok2 := target.(*PipelineParseError)
	return ok1 || ok2
}
