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

package metadata

import (
	"fmt"
	"strings"

	"github.com/drone/envsubst"
)

func EnvVarSubst(yaml string, environ map[string]string) (string, error) {
	return envsubst.Eval(yaml, func(name string) string {
		env := environ[name]
		// Quote values containing a YAML line break so a user-controlled value
		// (commit message, branch name, PR title, …) cannot break out of its
		// scalar and inject new YAML structure. YAML 1.1 treats LF, CR, NEL,
		// LS and PS as line breaks, so all of them must trigger quoting.
		if strings.ContainsAny(env, "\n\r\u0085  ") {
			env = fmt.Sprintf("%q", env)
		}
		return env
	})
}
