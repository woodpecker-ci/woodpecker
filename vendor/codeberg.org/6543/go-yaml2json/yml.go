package yaml2json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"gopkg.in/yaml.v3"
)

const maxDepth uint8 = 100

var (
	ErrBrokenMappingNode = errors.New("broken mapping node")
	ErrUnsupportedNode   = errors.New("unsupported yaml node")
	ErrMaxDepth          = errors.New("max depth reached")
)

// Convert YAML bytes to JSON bytes
func Convert(data []byte) ([]byte, error) {
	m := &yaml.Node{}
	if err := yaml.Unmarshal(data, m); err != nil {
		return nil, err
	}

	return ConvertNode(m)
}

// ConvertNode convert a gopkg.in/yaml.v3 Node to JSON bytes
func ConvertNode(m *yaml.Node) ([]byte, error) {
	n, err := resolveMerges(m)
	if err != nil {
		return nil, err
	}

	d, err := toJSON(n, 0)
	if err != nil {
		return nil, err
	}
	return json.Marshal(d)
}

// StreamConvert convert YAML byte stream to JSON byte stream
func StreamConvert(r io.Reader, w io.Writer) error {
	decoder := yaml.NewDecoder(r)
	encoder := json.NewEncoder(w)
	m := &yaml.Node{}

	if err := decoder.Decode(m); err != nil {
		return err
	}

	n, err := resolveMerges(m)
	if err != nil {
		return err
	}

	d, err := toJSON(n, 0)
	if err != nil {
		return err
	}

	return encoder.Encode(d)
}

// resolveMerges force yaml decoder to resolve map merges
func resolveMerges(m *yaml.Node) (*yaml.Node, error) {
	i := new(interface{})
	if err := m.Decode(i); err != nil {
		return nil, err
	}
	n := new(yaml.Node)
	return n, n.Encode(i)
}

// toJSON convert gopkg.in/yaml.v3 nodes to object that can be serialized as json
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func toJSON(node *yaml.Node, depth uint8) (interface{}, error) {
	// prevent loop by hardcoded limit
	if depth == maxDepth {
		return nil, ErrMaxDepth
	}

	switch node.Kind {
	case yaml.DocumentNode:
		return toJSON(node.Content[0], depth+1)

	case yaml.SequenceNode:
		val := make([]interface{}, len(node.Content))
		var err error
		for i := range node.Content {
			if val[i], err = toJSON(node.Content[i], depth+1); err != nil {
				return nil, err
			}
		}
		return val, nil

	case yaml.MappingNode:
		if (len(node.Content) % 2) != 0 {
			return nil, ErrBrokenMappingNode
		}
		val := make(map[string]interface{}, len(node.Content)%2)
		for i := len(node.Content); i > 1; i = i - 2 {
			k, err := toJSON(node.Content[i-2], depth+1)
			if err != nil {
				return nil, err
			}
			if val[fmt.Sprint(k)], err = toJSON(node.Content[i-1], depth+1); err != nil {
				return nil, err
			}
		}
		return val, nil

	case yaml.ScalarNode:
		switch node.Tag {
		case nullTag:
			return nil, nil
		case boolTag:
			return strconv.ParseBool(node.Value)
		case intTag:
			return strconv.ParseInt(node.Value, 10, 64)
		case floatTag:
			return strconv.ParseFloat(node.Value, 64)
		}
		return node.Value, nil

	case yaml.AliasNode:
		return toJSON(node.Alias, depth+1)
	}

	return nil, fmt.Errorf("%w: '%v'", ErrUnsupportedNode, node.Kind)
}

// Source: https://github.com/go-yaml/yaml/blob/3e3283e801afc229479d5fc68aa41df1137b8394/resolve.go#L70-L81
const (
	nullTag  = "!!null"
	boolTag  = "!!bool"
	intTag   = "!!int"
	floatTag = "!!float"
	// mergeTag     = "!!merge"     // TODO: do we have to parse this?
)
