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
	secretMapping := map[string]string{}
	assert.NoError(t, ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue, secretMapping))
	assert.EqualValues(t, want, got, "Problem converting plugin parameters to environment variables")

	// handle edge cases (#1609)
	got = map[string]string{}
	assert.NoError(t, ParamsToEnv(map[string]any{"a": []any{"a", nil}}, got, "PLUGIN_", true, nil, nil))
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

	assert.NoError(t, ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue, nil))
	assert.EqualValues(t, wantPrefixPlugin, got, "Problem converting plugin parameters to environment variables")

	wantNoPrefix := map[string]string{
		"STRING": "stringz",
		"INT":    "1",
	}

	// handle edge cases (#1609)
	got = map[string]string{}
	assert.NoError(t, ParamsToEnv(from, got, "", true, getSecretValue, nil))
	assert.EqualValues(t, wantNoPrefix, got, "Problem converting plugin parameters to environment variables")
}

func TestSanitizeParamKey(t *testing.T) {
	assert.EqualValues(t, "PLUGIN_DRY_RUN", sanitizeParamKey("PLUGIN_", true, "dry-run"))
	assert.EqualValues(t, "PLUGIN_DRY_RUN", sanitizeParamKey("PLUGIN_", true, "dry_Run"))
	assert.EqualValues(t, "PLUGIN_DRY_RUN", sanitizeParamKey("PLUGIN_", true, "dry.run"))
	assert.EqualValues(t, "PLUGIN_dry-run", sanitizeParamKey("PLUGIN_", false, "dry-run"))
	assert.EqualValues(t, "PLUGIN_dry_Run", sanitizeParamKey("PLUGIN_", false, "dry_Run"))
	assert.EqualValues(t, "PLUGIN_dry.run", sanitizeParamKey("PLUGIN_", false, "dry.run"))
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
	gotSecretMapping := map[string]string{}
	wantSecretMapping := map[string]string{
		"PLUGIN_MY_SECRET": "FooBar",
		"PLUGIN_MAP":       `{"entry2":["a","b",3],"key":"value","secret":"FooBar"}`,
		"PLUGIN_LIST_MAP":  `[{"password":"geheim","registry":"https://codeberg.org","username":"6543"}]`,
	}
	assert.NoError(t, ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue, gotSecretMapping))
	assert.Equal(t, wantSecretMapping, gotSecretMapping, "Problem collecting secret mapping")
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

	secretMapping := map[string]string{}
	assert.Error(t, ParamsToEnv(from, make(map[string]string), "PLUGIN_", true, getSecretValue, secretMapping))
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
	secretMapping := map[string]string{}
	assert.ErrorContains(t,
		ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue, secretMapping),
		fmt.Sprintf("secret %q not found or not allowed to be used", "secret_token"))
}

func TestSecretMappingSimpleSecret(t *testing.T) {
	from := map[string]any{
		"simple_secret": map[string]any{"from_secret": "my_token"},
		"regular_var":   "no_secret_here",
	}

	secrets := map[string]string{
		"my_token": "secret_value_123",
	}

	getSecretValue := func(name string) (string, error) {
		name = strings.ToLower(name)
		secret, ok := secrets[name]
		if ok {
			return secret, nil
		}
		return "", fmt.Errorf("secret %q not found", name)
	}

	got := map[string]string{}
	secretMapping := map[string]string{}

	assert.NoError(t, ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue, secretMapping))

	assert.Equal(t, "secret_value_123", got["PLUGIN_SIMPLE_SECRET"])
	assert.Equal(t, "no_secret_here", got["PLUGIN_REGULAR_VAR"])

	assert.Equal(t, "secret_value_123", secretMapping["PLUGIN_SIMPLE_SECRET"])
	assert.NotContains(t, secretMapping, "PLUGIN_REGULAR_VAR")
}

func TestSecretMappingComplexMapWithSecrets(t *testing.T) {
	from := map[string]any{
		"config": map[string]any{
			"database": map[string]any{
				"host":     "localhost",
				"password": map[string]any{"from_secret": "db_password"},
				"port":     5432,
			},
			"api_key": map[string]any{"from_secret": "api_secret"},
			"timeout": 30,
		},
		"simple_var": "no_secrets",
	}

	secrets := map[string]string{
		"db_password": "super_secret_db_pass",
		"api_secret":  "api_key_12345",
	}

	getSecretValue := func(name string) (string, error) {
		name = strings.ToLower(name)
		secret, ok := secrets[name]
		if ok {
			return secret, nil
		}
		return "", fmt.Errorf("secret %q not found", name)
	}

	got := map[string]string{}
	secretMapping := map[string]string{}

	assert.NoError(t, ParamsToEnv(from, got, "PLUGIN_", true, getSecretValue, secretMapping))

	expectedJSON := `{"api_key":"api_key_12345","database":{"host":"localhost","password":"super_secret_db_pass","port":5432},"timeout":30}`
	assert.Equal(t, expectedJSON, got["PLUGIN_CONFIG"])
	assert.Equal(t, "no_secrets", got["PLUGIN_SIMPLE_VAR"])

	assert.Equal(t, expectedJSON, secretMapping["PLUGIN_CONFIG"])
	assert.NotContains(t, secretMapping, "PLUGIN_SIMPLE_VAR")
}
