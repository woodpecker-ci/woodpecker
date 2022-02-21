package kubectl

// The run id
func (backend *KubeBackend) ID() string {
	return toKuberenetesValidName("wp-"+backend.activeRun.RunID, 30)
}

// The run namespace
func (backend *KubeBackend) Namespace() string {
	return backend.Client.CoreArgs.Namespace
}

func (backend *KubeBackend) DetachedJobs() []*KubeJobTemplate {
	return backend.activeRun.DetachedJobs
}

func (backend *KubeBackend) Context() string {
	return backend.Client.CoreArgs.Context
}
