package types

import "github.com/autom8ter/dagger/v3"

// Stage denotes a collection of one or more steps.
type Stage struct {
	Name  string  `json:"name,omitempty"`
	Alias string  `json:"alias,omitempty"`
	Steps []*Step `json:"steps,omitempty"`

	// TODO check if it can de-/serialized
	StepDAG dagger.GraphEdge[*Step]
}
