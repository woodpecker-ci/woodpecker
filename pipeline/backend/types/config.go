package types

// Config defines the runtime configuration of a pipeline.
type Config struct {
	Stages   []*Stage   `json:"pipeline"` // pipeline stages
	Networks []*Network `json:"networks"` // network definitions
	Volumes  []*Volume  `json:"volumes"`  // volume definitions
	Secrets  []*Secret  `json:"secrets"`  // secret definitions
}

// CliContext is the context key to pass cli context to backends if needed
var CliContext ContextKey

// ContextKey is just an empty struct. It exists so CliContext can be
// an immutable public variable with a unique type. It's immutable
// because nobody else can create a ContextKey, being unexported.
type ContextKey struct{}
