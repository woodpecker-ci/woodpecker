package yml

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// toJSON convert gopkg.in/yaml.v3 nodes to object that can be serialized as json
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func toJSON(node *yaml.Node) (interface{}, error) {
	switch node.Kind {
	case yaml.DocumentNode:
		return toJSON(node.Content[0])

	case yaml.SequenceNode:
		val := make([]interface{}, len(node.Content))
		var err error
		for i := range node.Content {
			if val[i], err = toJSON(node.Content[i]); err != nil {
				return nil, err
			}
		}
		return val, nil

	case yaml.MappingNode:
		if (len(node.Content) % 2) != 0 {
			return nil, fmt.Errorf("broken mapping node")
		}
		val := make(map[string]interface{}, len(node.Content)%2)
		for i := len(node.Content); i > 1; i = i - 2 {
			k, err := toJSON(node.Content[i-2])
			if err != nil {
				return nil, err
			}
			if val[fmt.Sprint(k)], err = toJSON(node.Content[i-1]); err != nil {
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
	}

	return nil, fmt.Errorf("do not support yaml node kind '%v'", node.Kind)
}

func ToJSON(data []byte) ([]byte, error) {
	m := &yaml.Node{}
	if err := yaml.Unmarshal(data, m); err != nil {
		return nil, err
	}

	d, err := toJSON(m)
	if err != nil {
		return nil, err
	}
	return json.Marshal(d)
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

// Source: https://github.com/go-yaml/yaml/blob/3e3283e801afc229479d5fc68aa41df1137b8394/resolve.go#L70-L81
const (
	nullTag  = "!!null"
	boolTag  = "!!bool"
	intTag   = "!!int"
	floatTag = "!!float"
	// strTag       = "!!str"       // we dont have to parse it
	// timestampTag = "!!timestamp" // TODO: do we have to parse this?
	// seqTag       = "!!seq"       // TODO: do we have to parse this?
	// mapTag       = "!!map"       // TODO: do we have to parse this?
	// binaryTag    = "!!binary"    // TODO: do we have to parse this?
	// mergeTag     = "!!merge"     // TODO: do we have to parse this?
)
