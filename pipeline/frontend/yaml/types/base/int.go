package base

import (
	"errors"
	"strconv"

	"github.com/docker/go-units"
)

// StringOrInt represents a string or an integer.
type StringOrInt int64

// UnmarshalYAML implements the Unmarshaler interface.
func (s *StringOrInt) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var intType int64
	if err := unmarshal(&intType); err == nil {
		*s = StringOrInt(intType)
		return nil
	}

	var stringType string
	if err := unmarshal(&stringType); err == nil {
		intType, err := strconv.ParseInt(stringType, 10, 64)
		if err != nil {
			return err
		}
		*s = StringOrInt(intType)
		return nil
	}

	return errors.New("Failed to unmarshal StringOrInt")
}

// MemStringOrInt represents a string or an integer
// the String supports notations like 10m for then Megabyte of memory
type MemStringOrInt int64

// UnmarshalYAML implements the Unmarshaler interface.
func (s *MemStringOrInt) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var intType int64
	if err := unmarshal(&intType); err == nil {
		*s = MemStringOrInt(intType)
		return nil
	}

	var stringType string
	if err := unmarshal(&stringType); err == nil {
		intType, err := units.RAMInBytes(stringType)
		if err != nil {
			return err
		}
		*s = MemStringOrInt(intType)
		return nil
	}

	return errors.New("Failed to unmarshal MemStringOrInt")
}
