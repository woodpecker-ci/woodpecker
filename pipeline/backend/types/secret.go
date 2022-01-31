package types

// Secret defines a runtime secret
type Secret struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Mask  bool   `json:"mask,omitempty"`
}
