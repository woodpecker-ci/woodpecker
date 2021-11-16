package types

import (
	"strconv"

	"gopkg.in/yaml.v3"
)

// BoolTrue is a custom Yaml boolean type that defaults to true.
type BoolTrue struct {
	value bool
}

// UnmarshalYAML implements custom Yaml unmarshaling.
func (b *BoolTrue) UnmarshalYAML(value *yaml.Node) error {
	var s string
	_ = value.Decode(&s)

	v, err := strconv.ParseBool(s)
	if err == nil {
		b.value = !v
	}
	if s != "" && err != nil {
		return err
	}
	return nil
}

// Bool returns the bool value.
func (b BoolTrue) Bool() bool {
	return !b.value
}
