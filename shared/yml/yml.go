package yml

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Source: https://github.com/icza/dyno/blob/f1bafe5d99965c48cc9d5c7cf024eeb495facc1e/dyno.go#L563-L601
// License: Apache 2.0 - Copyright 2017 Andras Belicza
// ConvertMapI2MapS walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type.
// This function comes handy if you want to marshal a dynamic object into
// JSON where maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func convertMapI2MapS(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = convertMapI2MapS(v2)
			default:
				m[fmt.Sprint(k)] = convertMapI2MapS(v2)
			}
		}
		v = m

	case []interface{}:
		for i, v2 := range x {
			x[i] = convertMapI2MapS(v2)
		}

	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = convertMapI2MapS(v2)
		}
	}

	return v
}

func ToJSON(data []byte) (j []byte, err error) {
	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	j, err = json.Marshal(convertMapI2MapS(m))
	if err != nil {
		return nil, err
	}

	return j, nil
}

func LoadYmlFileAsJSON(path string) (j []byte, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	j, err = ToJSON(data)
	if err != nil {
		return nil, err
	}

	return j, nil
}
