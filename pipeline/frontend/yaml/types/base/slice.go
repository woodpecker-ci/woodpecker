package base

import (
	"errors"
	"fmt"
)

// StringOrSlice represents a string or an array of strings.
// We need to override the yaml decoder to accept both options.
type StringOrSlice []string

// UnmarshalYAML implements the Unmarshaler interface.
func (s *StringOrSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringType string
	if err := unmarshal(&stringType); err == nil {
		*s = []string{stringType}
		return nil
	}

	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		parts, err := toStrings(sliceType)
		if err != nil {
			return err
		}
		*s = parts
		return nil
	}

	return errors.New("Failed to unmarshal StringOrSlice")
}

// MarshalYAML implements custom Yaml marshaling.
func (s StringOrSlice) MarshalYAML() (interface{}, error) {
	if len(s) == 0 {
		return "", nil
	} else if len(s) == 1 {
		return s[0], nil
	}
	return []string(s), nil
}

func toStrings(s []interface{}) ([]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	r := make([]string, len(s))
	for k, v := range s {
		if sv, ok := v.(string); ok {
			r[k] = sv
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
		}
	}
	return r, nil
}
