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

func NewSecretsReplacer(secrets []string) *strings.Replacer {
	var oldnew []string
	for _, old := range secrets {
		old = strings.TrimSpace(old)
		if len(old) == 0 {
			continue
		}
		// since replacer is executed on each line we have to split multi-line-secrets
		for _, part := range strings.Split(old, "\n") {
			if len(part) == 0 {
				continue
			}
			oldnew = append(oldnew, part)
			oldnew = append(oldnew, "********")
		}
	}

	return strings.NewReplacer(oldnew...)
}
