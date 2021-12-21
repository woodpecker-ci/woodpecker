package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestParamsToEnv(t *testing.T) {
	from := map[string]interface{}{
		"skip":         nil,
		"string":       "stringz",
		"int":          1,
		"float":        1.2,
		"bool":         true,
		"slice":        []int{1, 2, 3},
		"map":          map[string]string{"hello": "world"},
		"complex":      []struct{ Name string }{{"Jack"}, {"Jill"}},
		"complex2":     struct{ Name string }{"Jack"},
		"from.address": "noreply@example.com",
		"tags":         stringsToInterface("next", "latest"),
		"tag":          stringsToInterface("next"),
		"my_secret":    map[string]interface{}{"from_secret": "secret_token"},
	}
	want := map[string]string{
		"PLUGIN_STRING":       "stringz",
		"PLUGIN_INT":          "1",
		"PLUGIN_FLOAT":        "1.2",
		"PLUGIN_BOOL":         "true",
		"PLUGIN_SLICE":        "1,2,3",
		"PLUGIN_MAP":          `{"hello":"world"}`,
		"PLUGIN_COMPLEX":      `[{"name":"Jack"},{"name":"Jill"}]`,
		"PLUGIN_COMPLEX2":     `{"name":"Jack"}`,
		"PLUGIN_FROM_ADDRESS": "noreply@example.com",
		"PLUGIN_TAG":          "next",
		"PLUGIN_TAGS":         "next,latest",
		"PLUGIN_MY_SECRET":    "FooBar",
	}
	secrets := map[string]Secret{
		"secret_token": {Name: "secret_token", Value: "FooBar", Match: nil},
	}
	got := map[string]string{}
	assert.NoError(t, paramsToEnv(from, got, secrets))
	assert.EqualValues(t, want, got, "Problem converting plugin parameters to environment variables")
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
`)
	var from map[string]interface{}
	err := yaml.Unmarshal(fromYAML, &from)
	assert.NoError(t, err)

	want := map[string]string{
		"PLUGIN_STRING":    "stringz",
		"PLUGIN_INT":       "1",
		"PLUGIN_FLOAT":     "1.2",
		"PLUGIN_BOOL":      "true",
		"PLUGIN_SLICE":     "1,2,3",
		"PLUGIN_MY_SECRET": "FooBar",
	}
	secrets := map[string]Secret{
		"secret_token": {Name: "secret_token", Value: "FooBar", Match: nil},
	}
	got := map[string]string{}
	assert.NoError(t, paramsToEnv(from, got, secrets))
	assert.EqualValues(t, want, got, "Problem converting plugin parameters to environment variables")
}

func TestYAMLToParamsToEnvError(t *testing.T) {
	fromYAML := []byte(`my_secret:
  from_secret: not_a_secret
`)
	var from map[string]interface{}
	err := yaml.Unmarshal(fromYAML, &from)
	assert.NoError(t, err)
	secrets := map[string]Secret{
		"secret_token": {Name: "secret_token", Value: "FooBar", Match: nil},
	}
	assert.Error(t, paramsToEnv(from, make(map[string]string), secrets))
}

func stringsToInterface(val ...string) []interface{} {
	res := make([]interface{}, len(val))
	for i := range val {
		res[i] = val[i]
	}
	return res
}
