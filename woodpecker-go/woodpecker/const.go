package woodpecker

// Event values.
const (
	EventPush    = "push"
	EventPull    = "pull_request"
	EventTag     = "tag"
	EventDeploy  = "deployment"
	EventRelease = "release"
)

// Status values.
const (
	StatusBlocked = "blocked"
	StatusSkipped = "skipped"
	StatusPending = "pending"
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusKilled  = "killed"
	StatusError   = "error"
)
