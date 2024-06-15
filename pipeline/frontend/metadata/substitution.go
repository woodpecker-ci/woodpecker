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
	var missingEnvs []string
	out, err := envsubst.Eval(yaml, func(name string) string {
		env, has := environ[name]
		if !has {
			missingEnvs = append(missingEnvs, name)
		}
		if strings.Contains(env, "\n") {
			env = fmt.Sprintf("%q", env)
		}
		return env
	})
	if err != nil {
		return "", err
	}

	if len(missingEnvs) > 0 {
		return "", fmt.Errorf("missing env vars for substitution: %s", strings.Join(missingEnvs, ", "))
	}

	return out, nil
}
