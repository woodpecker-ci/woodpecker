package types

// Config defines the runtime configuration of a pipeline.
type Config struct {
	Stages   []*Stage   `json:"pipeline"` // pipeline stages
	Networks []*Network `json:"networks"` // network definitions
	Volumes  []*Volume  `json:"volumes"`  // volume definitions
	Secrets  []*Secret  `json:"secrets"`  // secret definitions
}
