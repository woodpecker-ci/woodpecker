package kubectl

// The run kubernetes id
func (run *KubePiplineRun) ID() string {
	return ToKuberenetesValidName("wp-"+run.RunID, 30)
}

// The kubernetes namespace
func (run *KubePiplineRun) Namespace() string {
	return run.Backend.Client.Namespace
}

// The kubernetes context
func (run *KubePiplineRun) Context() string {
	return run.Backend.Client.Context
}
