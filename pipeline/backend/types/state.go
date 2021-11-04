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

// // State defines the pipeline and process state.
// type State struct {
// 	Pipeline struct {
// 		// Current pipeline step
// 		Step *Step `json:"step"`
// 		// Current pipeline error state
// 		Error error `json:"error"`
// 	}
//
// 	Process struct {
// 		// Container exit code
// 		ExitCode int `json:"exit_code"`
// 		// Container exited, true or false
// 		Exited bool `json:"exited"`
// 		// Container is oom killed, true or false
// 		OOMKilled bool `json:"oom_killed"`
// 	}
// }
