package types

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type (
	// WorkflowNetworks defines a collection of networks.
	WorkflowNetworks struct {
		WorkflowNetworks []*WorkflowNetwork
	}

	// WorkflowNetwork defines a container network.
	WorkflowNetwork struct {
		Name       string            `yaml:"name,omitempty"`
		Driver     string            `yaml:"driver,omitempty"`
		DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (n *WorkflowNetworks) UnmarshalYAML(value *yaml.Node) error {
	networks := map[string]WorkflowNetwork{}
	err := value.Decode(&networks)

	for key, nn := range networks {
		if nn.Name == "" {
			nn.Name = fmt.Sprintf("%v", key)
		}
		if nn.Driver == "" {
			nn.Driver = "bridge"
		}
		n.WorkflowNetworks = append(n.WorkflowNetworks, &nn)
	}
	return err
}
