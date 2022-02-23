package kubectl

// The run kubernetes id
func (backend *KubeBackend) ID() string {
	return ToKuberenetesValidName("wp-"+backend.activeRun.RunID, 30)
}

// The kubernetes namespace
func (backend *KubeBackend) Namespace() string {
	return backend.Client.Namespace
}

// The kubernetes context
func (backend *KubeBackend) Context() string {
	return backend.Client.Context
}

// The current active run detached jobs
func (backend *KubeBackend) DetachedJobs() []*KubeJobTemplate {
	return backend.activeRun.DetachedJobs
}
