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

package types

// Config defines the runtime configuration of a pipeline.
type Config struct {
	Stages   []*Stage   `json:"pipeline"` // pipeline stages
	Networks []*Network `json:"networks"` // network definitions
	Volumes  []*Volume  `json:"volumes"`  // volume definitions
	Secrets  []*Secret  `json:"secrets"`  // secret definitions
}

// CliContext is the context key to pass cli context to backends if needed
var CliContext ContextKey

// ContextKey is just an empty struct. It exists so CliContext can be
// an immutable public variable with a unique type. It's immutable
// because nobody else can create a ContextKey, being unexported.
type ContextKey struct{}
