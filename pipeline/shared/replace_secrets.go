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

package shared

import "strings"

// NewSecretsReplacer creates a new strings.Replacer to replace sensitive
// strings with asterisks. It takes a slice of secrets strings as input
// and returns a populated strings.Replacer that will replace those
// secrets with asterisks. Each secret string is split on newlines to
// handle multi-line secrets.
func NewSecretsReplacer(secrets []string) *strings.Replacer {
	var oldNew []string

	// Strings shorter than minStringLength are not considered secrets.
	// Do not sanitize them.
	const minStringLength = 3

	for _, old := range secrets {
		old = strings.TrimSpace(old)
		if len(old) <= minStringLength {
			continue
		}
		// since replacer is executed on each line we have to split multi-line-secrets
		for _, part := range strings.Split(old, "\n") {
			if len(part) == 0 {
				continue
			}
			oldNew = append(oldNew, part)
			oldNew = append(oldNew, "********")
		}
	}

	return strings.NewReplacer(oldNew...)
}
