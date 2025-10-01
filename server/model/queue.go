package model

// QueueTask represents a task in the queue with additional API-specific fields.
type QueueTask struct {
	Task
	PipelineNumber int64  `json:"pipeline_number"`
	AgentName      string `json:"agent_name"`
}

// QueueInfo represents the response structure for queue information API.
type QueueInfo struct {
	Pending       []QueueTask `json:"pending"`
	WaitingOnDeps []QueueTask `json:"waiting_on_deps"`
	Running       []QueueTask `json:"running"`
	Stats         struct {
		WorkerCount        int `json:"worker_count"`
		PendingCount       int `json:"pending_count"`
		WaitingOnDepsCount int `json:"waiting_on_deps_count"`
		RunningCount       int `json:"running_count"`
	} `json:"stats"`
	Paused bool `json:"paused"`
} //	@name	QueueInfo
