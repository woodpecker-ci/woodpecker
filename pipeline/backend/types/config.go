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

// Config defines the runtime configuration of a workflow.
type Config struct {
	Stages  []*Stage  `json:"pipeline"` // workflow stages
	Network string    `json:"network"`  // network definition
	Volume  string    `json:"volume"`   // volume definition
	Secrets []*Secret `json:"secrets"`  // secret definitions
}

// CliCommand is the context key to pass cli context to backends if needed.
var CliCommand contextKey

// contextKey is just an empty struct. It exists so CliCommand can be
// an immutable public variable with a unique type. It's immutable
// because nobody else can create a ContextKey, being unexported.
type contextKey struct{}

// ImagePullOutput is an optional context key whose value is an
// io.Writer. When set, backends stream image-pull progress to it
// instead of os.Stdout. This lets an embedder (currently the CLI exec
// TUI; the agent could adopt it later) capture pull output so the
// docker client's progress writes cannot tear an alt-screen UI.
// Programmatic only: there is intentionally no flag for it.
//
// TODO: parse the jsonmessage stream and emit structured pull-progress
// log entries instead of raw text, so the web UI / TUI log panel can
// render it natively rather than as opaque lines.
var ImagePullOutput imagePullOutputKey

// imagePullOutputKey is a distinct unexported type so this context key
// cannot collide with contextKey (two empty-struct keys of the same
// type would compare equal and alias each other).
type imagePullOutputKey struct{}
