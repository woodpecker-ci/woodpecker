package types

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/strslice"
)

// Command represents a docker command, can be a string or an array of strings.
type Command strslice.StrSlice

// UnmarshalYAML implements the Unmarshaler interface.
func (s *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

	var interfaceType interface{}
	if err := unmarshal(&interfaceType); err == nil {
		fmt.Println(interfaceType)
	}

	return errors.New("failed to unmarshal Command")
}
