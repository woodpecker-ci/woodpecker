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
	"reflect"
	"strconv"
	"strings"

	"codeberg.org/6543/go-yaml2json"
	"gopkg.in/yaml.v3"
)

// ParamsToEnv uses reflection to convert a map[string]interface to a list
// of environment variables.
func ParamsToEnv(from map[string]interface{}, to, secrets map[string]string) (err error) {
	if to == nil {
		return fmt.Errorf("no map to write to")
	}
	for k, v := range from {
		if v == nil || len(k) == 0 {
			continue
		}
		to[sanitizeParamKey(k)], err = sanitizeParamValue(v, secrets)
		if err != nil {
			return err
		}
	}
	return nil
}

// format the environment variable key
func sanitizeParamKey(k string) string {
	return "PLUGIN_" + strings.ToUpper(
		strings.ReplaceAll(strings.ReplaceAll(k, ".", "_"), "-", "_"))
}

// indicate if a data type can be turned into string without encoding as json
func isComplex(t reflect.Kind) bool {
	switch t {
	case reflect.Bool,
		reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64:
		return false
	default:
		return true
	}
}

// sanitizeParamValue return the value of a setting as string prepared to be injected as environment variable
func sanitizeParamValue(v interface{}, secrets map[string]string) (string, error) {
	t := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)

	switch t.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(vv.Bool()), nil

	case reflect.String:
		return vv.String(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", vv.Int()), nil

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", vv.Float()), nil

	case reflect.Map:
		switch v := v.(type) {
		// gopkg.in/yaml.v3 only emit this map interface
		case map[string]interface{}:
			// check if it's a secret and return value if it's the case
			value, isSecret, err := injectSecret(v, secrets)
			if err != nil {
				return "", err
			} else if isSecret {
				return value, nil
			}
		default:
			return "", fmt.Errorf("could not handle: %#v", v)
		}

		return handleComplex(vv.Interface(), secrets)

	case reflect.Slice, reflect.Array:
		if vv.Len() == 0 {
			return "", nil
		}

		// if it's an interface unwrap and element check happen for each iteration later
		if t.Elem().Kind() == reflect.Interface ||
			// else check directly if element is not complex
			!isComplex(t.Elem().Kind()) {
			containComplex := false
			in := make([]string, vv.Len())

			for i := 0; i < vv.Len(); i++ {
				v := vv.Index(i).Interface()

				// ensure each element is not complex
				if isComplex(reflect.TypeOf(v).Kind()) {
					containComplex = true
					break
				}

				var err error
				if in[i], err = sanitizeParamValue(v, secrets); err != nil {
					return "", err
				}
			}

			if !containComplex {
				return strings.Join(in, ","), nil
			}
			return handleComplex(vv.Interface(), secrets)
		}
	}

	return handleComplex(vv.Interface(), secrets)
}

// handleComplex use yml.ToJSON to store configuration in environment variables
func handleComplex(v interface{}, secrets map[string]string) (string, error) {
	v, err := injectSecretRecursive(v, secrets)
	if err != nil {
		return "", err
	}

	out, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	out, err = yaml2json.Convert(out)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// injectSecret probe if map is actually a secret and so a string.
// if it's a string it returns either the value or an error if secret was not found
// else it just indicate to progress normally
func injectSecret(v map[string]interface{}, secrets map[string]string) (string, bool, error) {
	if secretNameI, ok := v["from_secret"]; ok {
		if secretName, ok := secretNameI.(string); ok {
			if secret, ok := secrets[strings.ToLower(secretName)]; ok {
				return secret, true, nil
			}
			return "", false, fmt.Errorf("no secret found for %q", secretName)
		}
		return "", false, fmt.Errorf("from_secret has to be a string")
	}
	return "", false, nil
}

// injectSecretRecursive iterates over all types and if they contain elements
// iterate recursively over them too, it uses injectSecret internally
func injectSecretRecursive(v interface{}, secrets map[string]string) (interface{}, error) {
	t := reflect.TypeOf(v)

	if !isComplex(t.Kind()) {
		return v, nil
	}

	switch t.Kind() {
	case reflect.Map:
		switch v := v.(type) {
		// gopkg.in/yaml.v3 only emit this map interface
		case map[string]interface{}:
			// handle secrets
			value, isSecret, err := injectSecret(v, secrets)
			if err != nil {
				return nil, err
			} else if isSecret {
				return value, nil
			}

			for key, val := range v {
				v[key], err = injectSecretRecursive(val, secrets)
				if err != nil {
					return nil, err
				}
			}
			return v, nil
		default:
			return v, fmt.Errorf("could not handle: %#v", v)
		}

	case reflect.Array, reflect.Slice:
		vv := reflect.ValueOf(v)
		vl := make([]interface{}, vv.Len())

		for i := 0; i < vv.Len(); i++ {
			v, err := injectSecretRecursive(vv.Index(i).Interface(), secrets)
			if err != nil {
				return nil, err
			}
			vl[i] = v
		}
		return vl, nil

	default:
		return v, nil
	}
}
