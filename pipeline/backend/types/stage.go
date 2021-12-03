package types

// Stage denotes a collection of one or more steps.
type Stage struct {
	Name  string  `json:"name,omitempty"`
	Alias string  `json:"alias,omitempty"`
	Steps []*Step `json:"steps,omitempty"`
}
