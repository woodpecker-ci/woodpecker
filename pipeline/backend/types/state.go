package types

// State defines a container state.
type State struct {
	// Container exit code
	ExitCode int `json:"exit_code"`
	// Container exited, true or false
	Exited bool `json:"exited"`
	// Container is oom killed, true or false
	OOMKilled bool `json:"oom_killed"`
}
