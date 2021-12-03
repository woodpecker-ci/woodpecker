package types

// Secret defines a runtime secret
type Secret struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Mount string `json:"mount,omitempty"`
	Mask  bool   `json:"mask,omitempty"`
}
