// Copyright 2025 Woodpecker Authors.
// Copyright 2024 "6543".
//
// Licensed under the MIT License.

package optional

import (
	"encoding/json"

	"go.yaml.in/yaml/v4"
)

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	var v *T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*o = FromPtr(v)
	return nil
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if !o.Has() {
		return []byte("null"), nil
	}

	return json.Marshal(o.Value())
}

func (o *Option[T]) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.DocumentNode && len(value.Content) == 1 {
		value = value.Content[0]
	}
	var v *T
	if err := value.Decode(&v); err != nil {
		return err
	}
	*o = FromPtr(v)
	return nil
}

func (o Option[T]) MarshalYAML() (any, error) {
	if !o.Has() {
		return nil, nil
	}

	value := new(yaml.Node)
	err := value.Encode(o.Value())
	return value, err
}
