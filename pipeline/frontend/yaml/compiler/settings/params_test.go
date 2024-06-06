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

package settings

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestParamsToEnv(t *testing.T) {
	from := map[string]any{
		"skip":             nil,
		"string":           "stringz",
		"int":              1,
		"float":            1.2,
		"bool":             true,
		"slice":            []int{1, 2, 3},
		"map":              map[string]any{"hello": "world"},
		"complex":          []struct{ Name string }{{"Jack"}, {"Jill"}},
		"complex2":         struct{ Name string }{"Jack"},
		"from.address":     "noreply@example.com",
		"tags":             stringsToInterface("next", "latest"),
		"tag":              stringsToInterface("next"),
		"my_secret":        map[string]any{"from_secret": "secret_token"},
		"UPPERCASE_SECRET": map[string]any{"from_secret": "SECRET_TOKEN"},
	}
	want := map[string]string{
		"PLUGIN_STRING":           "stringz",
		"PLUGIN_INT":              "1",
		"PLUGIN_FLOAT":            "1.2",
		"PLUGIN_BOOL":             "true",
		"PLUGIN_SLICE":            "1,2,3",
		"PLUGIN_MAP":              `{"hello":"world"}`,
		"PLUGIN_COMPLEX":          `[{"name":"Jack"},{"name":"Jill"}]`,
		"PLUGIN_COMPLEX2":         `{"name":"Jack"}`,
		"PLUGIN_FROM_ADDRESS":     "noreply@example.com",
		"PLUGIN_TAG":              "next",
		"PLUGIN_TAGS":             "next,latest",
		"PLUGIN_MY_SECRET":        "FooBar",
		"PLUGIN_UPPERCASE_SECRET": "FooBar",
	}
	secrets := map[string]string{
		"secret_token": "FooBar",
	}
	got := map[string]string{}
	getSecretValue := func(name string) (string, error) {
		name = strings.ToLower(name)
		secret, ok := secrets[name]
		if ok {
			return secret, nil
		}

		return "", fmt.Errorf("secret %q not found or not allowed to be used", name)
	}

	assert.NoError(t, ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue))
	assert.EqualValues(t, want, got, "Problem converting plugin parameters to environment variables")

	// handle edge cases (#1609)
	got = map[string]string{}
	assert.NoError(t, ParamsToEnv(map[string]any{"a": []any{"a", nil}}, got, "PLUGIN_", true, nil))
	assert.EqualValues(t, map[string]string{"PLUGIN_A": "a,"}, got)
}

func TestParamsToEnvPrefix(t *testing.T) {
	from := map[string]any{
		"string": "stringz",
		"int":    1,
	}
	wantPrefixPlugin := map[string]string{
		"PLUGIN_STRING": "stringz",
		"PLUGIN_INT":    "1",
	}
	got := map[string]string{}
	getSecretValue := func(name string) (string, error) {
		return "", fmt.Errorf("secret %q not found or not allowed to be used", name)
	}

	assert.NoError(t, ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue))
	assert.EqualValues(t, wantPrefixPlugin, got, "Problem converting plugin parameters to environment variables")

	wantNoPrefix := map[string]string{
		"STRING": "stringz",
		"INT":    "1",
	}

	// handle edge cases (#1609)
	got = map[string]string{}
	assert.NoError(t, ParamsToEnv(from, got, "", true, getSecretValue))
	assert.EqualValues(t, wantNoPrefix, got, "Problem converting plugin parameters to environment variables")
}

func TestSanitizeParamKey(t *testing.T) {
	assert.EqualValues(t, "PLUGIN_DRY_RUN", sanitizeParamKey("PLUGIN_", true, "dry-run"))
	assert.EqualValues(t, "PLUGIN_DRY_RUN", sanitizeParamKey("PLUGIN_", true, "dry_Run"))
	assert.EqualValues(t, "PLUGIN_DRY_RUN", sanitizeParamKey("PLUGIN_", true, "dry.run"))
	assert.EqualValues(t, "PLUGIN_dry_run", sanitizeParamKey("PLUGIN_", false, "dry-run"))
	assert.EqualValues(t, "PLUGIN_dry_Run", sanitizeParamKey("PLUGIN_", false, "dry_Run"))
}

func TestYAMLToParamsToEnv(t *testing.T) {
	fromYAML := []byte(`skip: ~
string: stringz
int: 1
float: 1.2
bool: true
slice: [1, 2, 3]
my_secret:
  from_secret: secret_token
map:
  key: "value"
  entry2:
    - "a"
    - "b"
    - 3
  secret:
    from_secret: secret_token
list.map:
  - registry: https://codeberg.org
    username: "6543"
    password:
      from_secret: cb_password
`)
	var from map[string]any
	err := yaml.Unmarshal(fromYAML, &from)
	assert.NoError(t, err)

	want := map[string]string{
		"PLUGIN_STRING":    "stringz",
		"PLUGIN_INT":       "1",
		"PLUGIN_FLOAT":     "1.2",
		"PLUGIN_BOOL":      "true",
		"PLUGIN_SLICE":     "1,2,3",
		"PLUGIN_MY_SECRET": "FooBar",
		"PLUGIN_MAP":       `{"entry2":["a","b",3],"key":"value","secret":"FooBar"}`,
		"PLUGIN_LIST_MAP":  `[{"password":"geheim","registry":"https://codeberg.org","username":"6543"}]`,
	}
	secrets := map[string]string{
		"secret_token": "FooBar",
		"cb_password":  "geheim",
	}
	got := map[string]string{}
	getSecretValue := func(name string) (string, error) {
		name = strings.ToLower(name)
		secret, ok := secrets[name]
		if ok {
			return secret, nil
		}

		return "", fmt.Errorf("secret %q not found or not allowed to be used", name)
	}

	assert.NoError(t, ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue))
	assert.EqualValues(t, want, got, "Problem converting plugin parameters to environment variables")
}

func TestYAMLToParamsToEnvError(t *testing.T) {
	fromYAML := []byte(`my_secret:
  from_secret: not_a_secret
`)
	var from map[string]any
	err := yaml.Unmarshal(fromYAML, &from)
	assert.NoError(t, err)
	secrets := map[string]string{
		"secret_token": "FooBar",
	}
	getSecretValue := func(name string) (string, error) {
		name = strings.ToLower(name)
		secret, ok := secrets[name]
		if ok {
			return secret, nil
		}

		return "", fmt.Errorf("secret %q not found or not allowed to be used", name)
	}

	assert.Error(t, ParamsToEnv(from, make(map[string]string), "PLUGIN_", true, getSecretValue))
}

func stringsToInterface(val ...string) []any {
	res := make([]any, len(val))
	for i := range val {
		res[i] = val[i]
	}
	return res
}

func TestSecretNotFound(t *testing.T) {
	from := map[string]any{
		"map": map[string]any{"secret": map[string]any{"from_secret": "secret_token"}},
	}

	secrets := map[string]string{
		"a_different_password": "secret",
	}
	getSecretValue := func(name string) (string, error) {
		name = strings.ToLower(name)
		secret, ok := secrets[name]
		if ok {
			return secret, nil
		}

		return "", fmt.Errorf("secret %q not found or not allowed to be used", name)
	}
	got := map[string]string{}

	assert.ErrorContains(t,
		ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue),
		fmt.Sprintf("secret %q not found or not allowed to be used", "secret_token"))
}
