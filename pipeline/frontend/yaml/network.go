package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type (
	// Networks defines a collection of networks.
	Networks struct {
		Networks []*Network
	}

	// Network defines a container network.
	Network struct {
		Name       string            `yaml:"name,omitempty"`
		Driver     string            `yaml:"driver,omitempty"`
		DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (n *Networks) UnmarshalYAML(value *yaml.Node) error {
	networks := map[string]Network{}
	err := value.Decode(&networks)

	for key, nn := range networks {
		if nn.Name == "" {
			nn.Name = fmt.Sprintf("%v", key)
		}
		if nn.Driver == "" {
			nn.Driver = "bridge"
		}
		n.Networks = append(n.Networks, &nn)
	}
	return err
}
