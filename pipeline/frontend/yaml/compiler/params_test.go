package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	}
	got := map[string]string{}
	assert.NoError(t, paramsToEnv(from, got))
	assert.EqualValues(t, want, got, "Problem converting plugin parameters to environment variables")
}

func stringsToInterface(val ...string) []interface{} {
	res := make([]interface{}, len(val))
	for i := range val {
		res[i] = val[i]
	}
	return res
}
