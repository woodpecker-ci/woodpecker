// Copyright 2026 Woodpecker Authors
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

package settings

import (
	"testing"

	"go.yaml.in/yaml/v4"
)

// FuzzParamsToEnv drives the reflection and recursion heavy plugin settings
// conversion (incl. from_secret injection) with untrusted structures decoded
// from yaml and checks that it never panics.
func FuzzParamsToEnv(f *testing.F) {
	f.Add("string: stringz\nint: 1\nfloat: 1.2\nbool: true")
	f.Add("slice: [1, 2, 3]\nmap: { hello: world }")
	f.Add("my_secret:\n  from_secret: secret_token")
	f.Add("nested:\n  - a:\n      from_secret: tok\n  - b: [x, {c: d}]")

	getSecret := func(name string) (string, error) {
		return "secret_" + name, nil
	}

	f.Fuzz(func(_ *testing.T, data string) {
		from := map[string]any{}
		if err := yaml.Unmarshal([]byte(data), &from); err != nil {
			return
		}
		to := map[string]string{}
		_ = ParamsToEnv(from, to, "PLUGIN_", true, getSecret, nil)
	})
}
