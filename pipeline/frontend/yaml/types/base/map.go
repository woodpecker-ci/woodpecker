package base

import (
	"errors"
	"fmt"
	"strings"
)

// SliceOrMap represents a map of strings, string slice are converted into a map
type SliceOrMap map[string]string

// UnmarshalYAML implements the Unmarshaler interface.
func (s *SliceOrMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		parts := map[string]string{}
		for _, s := range sliceType {
			if str, ok := s.(string); ok {
				str := strings.TrimSpace(str)
				keyValueSlice := strings.SplitN(str, "=", 2)

				key := keyValueSlice[0]
				val := ""
				if len(keyValueSlice) == 2 {
					val = keyValueSlice[1]
				}
				parts[key] = val
			} else {
				return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", s, s)
			}
		}
		*s = parts
		return nil
	}

	var mapType map[interface{}]interface{}
	if err := unmarshal(&mapType); err == nil {
		parts := map[string]string{}
		for k, v := range mapType {
			if sk, ok := k.(string); ok {
				if sv, ok := v.(string); ok {
					parts[sk] = sv
				} else {
					return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
				}
			} else {
				return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", k, k)
			}
		}
		*s = parts
		return nil
	}

	return errors.New("Failed to unmarshal SliceOrMap")
}
