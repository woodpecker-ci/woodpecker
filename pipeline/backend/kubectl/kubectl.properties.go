package kubectl

// The run id
func (backend *KubeCtlBackend) ID() string {
	return "wp-" + backend.RunID
}

// The run namespace
func (backend *KubeCtlBackend) Namespace() string {
	return backend.Client.CoreArgs.Namespace
}
