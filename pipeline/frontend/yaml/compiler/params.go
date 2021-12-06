package compiler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/shared/yml"
)

// paramsToEnv uses reflection to convert a map[string]interface to a list
// of environment variables.
func paramsToEnv(from map[string]interface{}, to map[string]string) (err error) {
	if to == nil {
		return fmt.Errorf("no map to write to")
	}
	for k, v := range from {
		if v == nil || len(k) == 0 {
			continue
		}
		to[sanitizeParamKey(k)], err = sanitizeParamValue(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func sanitizeParamKey(k string) string {
	return "PLUGIN_" +
		strings.ToUpper(
			strings.ReplaceAll(k, ".", "_"),
		)
}

func isComplex(t reflect.Kind) bool {
	switch t {
	case reflect.Bool,
		reflect.String,
		reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64:
		return false
	default:
		return true
	}
}

func sanitizeParamValue(v interface{}) (string, error) {
	t := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)

	switch t.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(vv.Bool()), nil

	case reflect.String:
		return vv.String(), nil

	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return fmt.Sprintf("%v", vv.Int()), nil

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", vv.Float()), nil

	case reflect.Map:
		ymlOut, _ := yaml.Marshal(vv.Interface())
		out, _ := yml.ToJSON(ymlOut)
		return string(out), nil

	case reflect.Slice, reflect.Array:
		if !isComplex(t.Elem().Kind()) {
			in := make([]string, vv.Len())
			for i := 0; i < vv.Len(); i++ {
				var err error
				if in[i], err = sanitizeParamValue(vv.Index(i).Interface()); err != nil {
					return "", err
				}
			}
			return strings.Join(in, ","), nil
		}

		// it's complex use yml.ToJSON
		fallthrough

	default:
		out, err := yaml.Marshal(vv.Interface())
		if err != nil {
			return "", err
		}
		out, err = yml.ToJSON(out)
		if err != nil {
			return "", err
		}
		return string(out), nil
	}
}
