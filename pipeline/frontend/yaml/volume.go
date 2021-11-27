package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type (
	// Volumes defines a collection of volumes.
	Volumes struct {
		Volumes []*Volume
	}

	// Volume defines a container volume.
	Volume struct {
		Name       string            `yaml:"name,omitempty"`
		Driver     string            `yaml:"driver,omitempty"`
		DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (v *Volumes) UnmarshalYAML(value *yaml.Node) error {
	y, _ := yaml.Marshal(value)

	volumes := map[string]Volume{}
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
		v.Volumes = append(v.Volumes, &vv)
	}
	return err
}
