package compiler

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func TestParamsToEnv(t *testing.T) {
	from := map[string]interface{}{
		"skip":    nil,
		"string":  "stringz",
		"int":     1,
		"float":   1.2,
		"bool":    true,
		"map":     map[string]string{"hello": "world"},
		"slice":   []int{1, 2, 3},
		"complex": []struct{ Name string }{{"Jack"}, {"Jill"}},
	}
	want := map[string]string{
		"PLUGIN_STRING":  "stringz",
		"PLUGIN_INT":     "1",
		"PLUGIN_FLOAT":   "1.2",
		"PLUGIN_BOOL":    "true",
		"PLUGIN_MAP":     `{"hello":"world"}`,
		"PLUGIN_SLICE":   "1,2,3",
		"PLUGIN_COMPLEX": `[{"name":"Jack"},{"name":"Jill"}]`,
	}
	got := map[string]string{}
	paramsToEnv(from, got)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Problem converting plugin parameters to environment variables")
		pretty.Ldiff(t, want, got)
	}
}
