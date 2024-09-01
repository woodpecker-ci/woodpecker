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

package linter

// Option configures a linting option.
type Option func(*Linter)

// WithTrusted adds the trusted option to the linter.
func WithTrusted(trusted bool) Option {
	return func(linter *Linter) {
		linter.trusted = trusted
	}
}

// PrivilegedPlugins adds the list of privileged plugins.
func PrivilegedPlugins(plugins []string) Option {
	return func(linter *Linter) {
		linter.privilegedPlugins = &plugins
	}
}

// WithTrustedClonePlugins adds the list of trusted clone plugins.
func WithTrustedClonePlugins(plugins []string) Option {
	return func(linter *Linter) {
		linter.trustedClonePlugins = &plugins
	}
}
