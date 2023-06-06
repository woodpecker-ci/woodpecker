package types

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type (
	// WorkflowVolumes defines a collection of volumes.
	WorkflowVolumes struct {
		WorkflowVolumes []*WorkflowVolume
	}

	// WorkflowVolume defines a container volume.
	WorkflowVolume struct {
		Name       string            `yaml:"name,omitempty"`
		Driver     string            `yaml:"driver,omitempty"`
		DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (v *WorkflowVolumes) UnmarshalYAML(value *yaml.Node) error {
	y, _ := yaml.Marshal(value)

	volumes := map[string]WorkflowVolume{}
	err := yaml.Unmarshal(y, &volumes)
	if err != nil {
		return err
	}

	for key, vv := range volumes {
		if vv.Name == "" {
			vv.Name = fmt.Sprintf("%v", key)
		}
		if vv.Driver == "" {
			vv.Driver = "local"
		}
		v.WorkflowVolumes = append(v.WorkflowVolumes, &vv)
	}
	return err
}
