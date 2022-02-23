package kubectl

// The run kubernetes id
func (run *KubeBackendRun) ID() string {
	return ToKuberenetesValidName("wp-"+run.RunID, 30)
}

// The kubernetes namespace
func (run *KubeBackendRun) Namespace() string {
	return run.Backend.Client.Namespace
}

// The kubernetes context
func (run *KubeBackendRun) Context() string {
	return run.Backend.Client.Context
}
