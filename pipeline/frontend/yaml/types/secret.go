package types

import "gopkg.in/yaml.v3"

type (
	// Secrets defines a collection of secrets.
	Secrets struct {
		Secrets []*Secret
	}

	// Secret defines a container secret.
	Secret struct {
		Source string `yaml:"source"`
		Target string `yaml:"target"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (s *Secrets) UnmarshalYAML(value *yaml.Node) error {
	y, _ := yaml.Marshal(value)

	var strslice []string
	err := yaml.Unmarshal(y, &strslice)
	if err == nil {
		for _, str := range strslice {
			s.Secrets = append(s.Secrets, &Secret{
				Source: str,
				Target: str,
			})
		}
		return nil
	}
	return yaml.Unmarshal(y, &s.Secrets)
}
